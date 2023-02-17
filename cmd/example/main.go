package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

// Fill in you partner api key and public key pem here
const partnerAPIKey = ""
const rsaPrivateKey = ``

type qredoRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

type qredoRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

type httpRequest struct {
	url       string
	method    string
	body      []byte
	timestamp string
}

func main() {
	store := util.NewFileStore("examplestore.db")
	err := store.Init()
	if err != nil {
		panic(errors.Wrap(err, "file store init"))
	}
	cfg := config.Config{
		Base: config.Base{
			QredoAPI: "https://play-api.qredo.network/api/v1/p",
			PIN:      0,
		},
	}

	core, err := lib.New(&cfg, store)
	if err != nil {
		panic(err)
	}

	// initiate on-boarding of new agent
	regInitResponse, err := core.ClientRegister("agent-test1")
	if err != nil {
		panic(errors.Wrap(err, "register init"))
	}

	// send EC and BLS keys to Qredo Partner API, initializing the new agent in the backend
	qredoReq := &qredoRegisterInitRequest{
		Name:         "agent-test",
		BLSPublicKey: regInitResponse.BLSPublicKey,
		ECPublicKey:  regInitResponse.ECPublicKey,
	}
	body, err := json.Marshal(qredoReq)
	if err != nil {
		panic(err)
	}
	req := &httpRequest{
		timestamp: fmt.Sprintf("%v", time.Now().Unix()),
		method:    "POST",
		url:       fmt.Sprintf("%s/coreclient/init", cfg.Base.QredoAPI),
		body:      body,
	}
	signature := partnerSign(req)
	header := http.Header{}
	header.Add("x-api-key", partnerAPIKey)
	header.Add("x-sign", signature)
	header.Add("x-timestamp", req.timestamp)
	qredoResponse := &qredoRegisterInitResponse{}
	httpClient := util.NewHTTPClient()
	err = httpClient.Request(req.method, req.url, qredoReq, qredoResponse, header)
	if err != nil {
		panic(errors.Wrap(err, "request to partner api"))
	}

	// finish registration of agent.
	// after this step there will be an entry in examplestore.db for the new agent
	finishResp, err := core.ClientRegisterFinish(&api.ClientRegisterFinishRequest{
		ID:           qredoResponse.ID,
		ClientID:     qredoResponse.ClientID,
		ClientSecret: qredoResponse.ClientSecret,
		AccountCode:  qredoResponse.AccountCode,
		IDDocument:   qredoResponse.IDDocument,
	}, regInitResponse.RefID)
	if err != nil {
		panic(errors.Wrap(err, "client register finish request"))
	}

	fmt.Printf("created agent\nid: %s\nfeed url: %s\n", qredoResponse.ID, finishResp.FeedURL)
}

func partnerSign(req *httpRequest) string {
	h := sha256.New()
	h.Write([]byte(req.timestamp))
	h.Write([]byte(req.url))
	h.Write(req.body)
	digest := h.Sum(nil)

	block, _ := pem.Decode([]byte(rsaPrivateKey))

	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(errors.Wrap(err, "parse RSA key"))
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, digest)
	if err != nil {
		panic(errors.Wrap(err, "sign request"))
	}

	return base64.RawURLEncoding.EncodeToString(signature)
}
