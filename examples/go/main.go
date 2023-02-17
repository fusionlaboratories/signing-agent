package main

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	agent_name := "test-agent"
	rsa_key_file := "private.pem"
	api_key := os.Getenv("APIKEY")
	company_id := os.Getenv("CUSTOMERID")
	host := "127.0.0.1"
	port := 8007

	trx_callback := func(msg TransactionMessage) bool {
		msg_json, _ := json.MarshalIndent(msg, "", "  ")
		log.Println(string(msg_json))

		status_details := msg.Details["statusDetails"].(map[string]interface{})
		net_amount := status_details["netAmount"].(float64)
		log.Println(net_amount)

		if net_amount < 100_000 {
			// approve
			log.Println("Approving transaction")
			return true
		}

		// ... more conditions

		// reject
		log.Println("Rejecting transaction")
		return false
	}

	client := NewSigingAgentClient(agent_name, rsa_key_file, api_key, company_id, host, port, trx_callback)
	agent_id, err := client.Init()
	if err != nil {
		log.Panicf("error while initializing agent: %s", err)
	}

	log.Println(agent_id)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
}

type SigningAgentClient struct {
	agent_name           string
	rsa_key_pem          []byte
	rsa_key              *rsa.PrivateKey
	api_key              string
	company_id           string
	host                 string
	port                 int
	transaction_callback func(TransactionMessage) bool
}

type TransactionMessage struct {
	Id           string `json:"id"`
	TrxType      string `json:"type"`
	CoreClientId string `json:"coreClientID"`
	Status       string `json:"status"`
	Timestamp    int    `json:"timestamp"`
	ExpireTime   int    `json:"expireTime"`
	Details      map[string]interface{}
}

func NewSigingAgentClient(agent_name, rsa_key_file, api_key, company_id, host string, port int, transaction_callback func(TransactionMessage) bool) *SigningAgentClient {
	rsa_key_pem, err := os.ReadFile(rsa_key_file)
	if err != nil {
		log.Fatalf("error while reading rsa key pem file: %s\n", err)
	}

	decoded_pem, _ := pem.Decode(rsa_key_pem)
	rsa_key, err := x509.ParsePKCS1PrivateKey(decoded_pem.Bytes)
	if err != nil {
		log.Fatalf("error while parsing rsa key pem: %s\n", err)
	}

	return &SigningAgentClient{
		agent_name:           agent_name,
		rsa_key_pem:          rsa_key_pem,
		rsa_key:              rsa_key,
		api_key:              api_key,
		company_id:           company_id,
		host:                 host,
		port:                 port,
		transaction_callback: transaction_callback,
	}
}

func (c *SigningAgentClient) Init() (string, error) {
	agent_id, err := c.registerAgent()
	if err == nil {
		go c.connectFeed()
	}

	return agent_id, err
}

func (c *SigningAgentClient) registerAgent() (string, error) {
	agent_id_result := []string{}
	err := c.executeAgentApiCall("GET", fmt.Sprintf("http://%s:%d/api/v1/client", c.host, c.port), nil, &agent_id_result)
	if err != nil {
		return "", err
	}

	if len(agent_id_result) > 0 {
		return agent_id_result[0], nil
	}

	b64pem := base64.StdEncoding.EncodeToString(c.rsa_key_pem)
	register_result := make(map[string]interface{}, 0)
	err = c.executeAgentApiCall("POST", fmt.Sprintf("http://%s:%d/api/v1/register", c.host, c.port), map[string]interface{}{
		"name":             c.agent_name,
		"APIKey":           c.api_key,
		"base64PrivateKey": b64pem,
	}, &register_result)
	if err != nil {
		return "", err
	}

	return register_result["agentId"].(string), nil
}

func (c *SigningAgentClient) connectFeed() error {

	done := make(chan struct{})
	close_handler := func(code int, text string) error {
		log.Printf("feed disconnected: %s\n", text)
		close(done)
		<-time.After(5 * time.Second)
		go c.connectFeed()
		return nil
	}

	w, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%d/api/v1/client/feed", c.host, c.port), nil)
	if err != nil {
		log.Println("dial:", err)
		return close_handler(-1, err.Error())
	}
	defer w.Close()

	log.Println("feed connected")

	w.SetCloseHandler(close_handler)

	go func() {
		defer func() {
			_, ok := <-done
			if ok {
				close(done)
			}
		}()

		for {
			_, message, err := w.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				close_handler(-1, err.Error())
				return
			}

			msg := TransactionMessage{}
			err = json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("unmarshal:", err)
				return
			}

			details, err := c.getTransactionDetails(msg)
			if err != nil {
				log.Println("getTransactionDetails:", err)
			}

			msg.Details = details

			if c.transaction_callback(msg) {
				c.approveTransaction(msg.Id)
			} else {
				c.rejectTransaction(msg.Id)
			}
		}
	}()

	for {
		select {
		case <-done:
			return nil
		default:
			<-time.After(10 * time.Millisecond)
		}
	}
}

func (c *SigningAgentClient) approveTransaction(transaction_id string) error {
	return c.executeAgentApiCall("PUT", fmt.Sprintf("http://%s:%d/api/v1/client/action/%s", c.host, c.port, transaction_id), nil, nil)
}

func (c *SigningAgentClient) rejectTransaction(transaction_id string) error {
	return c.executeAgentApiCall("DELETE", fmt.Sprintf("http://%s:%d/api/v1/client/action/%s", c.host, c.port, transaction_id), nil, nil)
}

func (c *SigningAgentClient) getTransactionDetails(msg TransactionMessage) (map[string]interface{}, error) {
	if len(c.company_id) < 1 {
		return nil, nil
	}

	type_url := ""
	if msg.TrxType == "ApproveWithdraw" {
		type_url = "withdraw"
	} else {
		type_url = "transfer"
	}

	result := map[string]interface{}{}
	err := c.executePartnerApiCall("GET", fmt.Sprintf("https://play-api.qredo.network/api/v1/p/company/%s/%s/%s", c.company_id, type_url, msg.Id), nil, &result)
	return result, err
}

func (c *SigningAgentClient) executeAgentApiCall(method string, url string, body map[string]interface{}, res interface{}) error {
	var body_reader io.Reader = nil
	if len(body) > 0 {
		body_json, _ := json.Marshal(body)
		body_reader = bytes.NewBuffer(body_json)
	}

	req, err := http.NewRequest(method, url, body_reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("non 200 status code received: %s %s -> %s", method, url, response.Status)
	}

	if res != nil {
		if err := json.NewDecoder(response.Body).Decode(res); err != nil {
			return err
		}
	}

	return nil
}

func (c *SigningAgentClient) executePartnerApiCall(method string, url string, body map[string]interface{}, res interface{}) error {
	var body_json []byte
	var body_reader io.Reader = nil
	if len(body) > 0 {
		body_json, _ = json.Marshal(body)
		body_reader = bytes.NewBuffer(body_json)
	}

	req, err := http.NewRequest(method, url, body_reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("x-api-key", c.api_key)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	req.Header.Add("x-timestamp", timestamp)

	sig, err := c.signRequest(url, timestamp, body_json)
	if err != nil {
		return err
	}
	req.Header.Add("x-sign", sig)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("non 200 status code received: %s %s -> %s", method, url, response.Status)
	}

	if res != nil {
		if err := json.NewDecoder(response.Body).Decode(res); err != nil {
			return err
		}
	}

	return nil
}

func (c *SigningAgentClient) signRequest(uri, timestamp string, body []byte) (string, error) {
	h := sha256.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(uri))
	if body != nil {
		h.Write(body)
	}

	dgst := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(nil, c.rsa_key, crypto.SHA256, dgst)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(signature), nil
}
