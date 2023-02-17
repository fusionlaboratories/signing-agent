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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

type SaCli struct {
	Agent         lib.SigningAgentClient
	QredoAPI      string
	PartnerAPIKey string
	PrivateKey    string
}

func NewDemo(url, apiKey, privateKey string) (*SaCli, error) {
	store := util.NewFileStore("demo.db")
	if err := store.Init(); err != nil {
		return nil, errors.Wrap(err, "file store init")
	}
	cfg := config.Config{
		Base: config.Base{
			PIN:      123,
			QredoAPI: url,
		},
	}

	agent, err := lib.New(&cfg, store)
	if err != nil {
		return nil, err
	}

	demo := &SaCli{
		Agent:         agent,
		QredoAPI:      cfg.Base.QredoAPI,
		PartnerAPIKey: apiKey,
		PrivateKey:    privateKey,
	}
	return demo, nil
}

func (d *SaCli) Register(name string) (*ClientRegisterResponse, error) {
	clRegResp, err := d.Agent.ClientRegister("demo-agent")
	if err != nil {
		return nil, errors.Wrap(err, "client register")
	}

	qredoReq := &ClientRegisterInitRequest{
		Name:         name,
		BLSPublicKey: clRegResp.BLSPublicKey,
		ECPublicKey:  clRegResp.ECPublicKey,
	}
	body, err := json.Marshal(qredoReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/coreclient/init", d.QredoAPI)
	req := &httpRequest{timestamp: fmt.Sprintf("%v", time.Now().Unix()), method: "POST", url: url, body: body}

	signature, err := d.sign(req.body, req.timestamp, req.url)
	if err != nil {
		return nil, err
	}

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)
	header.Add("x-sign", signature)
	header.Add("x-timestamp", req.timestamp)

	crResp := &ClientRegisterInitResponse{}
	httpClient := util.NewHTTPClient()
	if err = httpClient.Request(req.method, req.url, qredoReq, crResp, header); err != nil {
		return nil, errors.Wrap(err, "request to partner api")
	}

	finishResp, err := d.Agent.ClientRegisterFinish(&api.ClientRegisterFinishRequest{
		ID:           crResp.ID,
		ClientID:     crResp.ClientID,
		ClientSecret: crResp.ClientSecret,
		AccountCode:  crResp.AccountCode,
		IDDocument:   crResp.IDDocument,
	}, clRegResp.RefID)
	if err != nil {
		return nil, errors.Wrap(err, "client register finish request")
	}

	resp := &ClientRegisterResponse{}
	resp.ID = crResp.ID
	resp.ClientID = crResp.ClientID
	resp.AccountCode = crResp.AccountCode
	resp.ClientSecret = crResp.ClientSecret
	resp.IDDocument = crResp.IDDocument
	resp.Timestamp = crResp.Timestamp
	resp.FeedURL = finishResp.FeedURL

	return resp, nil
}

func (d *SaCli) Approve(actionID string) error {
	if err := d.Agent.ActionApprove(actionID); err != nil {
		return err
	}
	return nil
}

func (d *SaCli) CreateCompany(name, city, country, domain, ref string) (*CreateCompanyResponse, error) {
	reqCC := &CreateCompanyRequest{}
	reqCC.Name = name
	reqCC.City = city
	reqCC.Country = country
	reqCC.Domain = domain
	reqCC.Ref = ref

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	ResCC := &CreateCompanyResponse{}
	url := fmt.Sprintf("%s/company", d.QredoAPI)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("POST", url, reqCC, ResCC, header); err != nil {
		return nil, errors.Wrap(err, "request to partner api")
	}
	return ResCC, nil
}

func (d *SaCli) AddTrustedparty(companyID, agentId string) error {
	reqAtp := &AddTrustedPartyRequest{Address: agentId}

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	url := fmt.Sprintf("%s/company/%s/trustedparty", d.QredoAPI, companyID)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("POST", url, reqAtp, nil, header); err != nil {
		return errors.Wrap(err, "request to partner api")
	}
	return nil
}

func (d *SaCli) CreateFund(companyID, fundName, fundDesc, memberID string) (*AddFundResponse, error) {
	reqAF := &AddFundRequest{
		FundName:        fundName,
		FundDescription: fundDesc,
		CustodygroupWithdraw: CustodyGroup{
			Threshold: 1,
			Members:   []string{memberID},
		},
		CustodygroupTx: CustodyGroup{
			Threshold: 1,
			Members:   []string{memberID},
		},
		Wallets: []FundWallet{
			{
				Name:  "New wallet with custom custody group",
				Asset: "ETH-GOERLI",
				CustodygroupWithdraw: &CustodyGroup{
					Threshold: 1,
					Members:   []string{memberID},
				},
				CustodygroupTx: &CustodyGroup{
					Threshold: 1,
					Members:   []string{memberID},
				},
			},
		},
	}

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	fRes := &AddFundResponse{}
	url := fmt.Sprintf("%s/company/%s/fund", d.QredoAPI, companyID)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("POST", url, reqAF, fRes, header); err != nil {
		return nil, errors.Wrap(err, "request to partner api")
	}
	return fRes, nil
}

func (d *SaCli) AddWhitelist(companyID, fundID, address string) error {
	reqAWL := &AddWhitelistRequest{
		Address: address,
		Asset:   "ETH-GOERLI",
		Name:    "Metamask",
	}

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	url := fmt.Sprintf("%s/company/%s/fund/%s/whitelist", d.QredoAPI, companyID, fundID)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("POST", url, reqAWL, nil, header); err != nil {
		return errors.Wrap(err, "request to partner api")
	}
	return nil
}

func (d *SaCli) GetDepositList(companyID, fundID string) (*DepositAddressListResponse, error) {
	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	respDa := &DepositAddressListResponse{}
	url := fmt.Sprintf("%s/company/%s/fund/%s/deposit", d.QredoAPI, companyID, fundID)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("GET", url, nil, respDa, header); err != nil {
		return nil, errors.Wrap(err, "request to partner api")
	}
	return respDa, nil
}

func (d *SaCli) Withdraw(companyID, walletID, address string, amount int64) (*NewTransactionResponse, error) {
	reqWD := &NewWithdrawRequest{
		WalletID: walletID,
		Address:  address,
		Send: AssetAmount{
			Asset:  "ETH-GOERLI",
			Amount: amount,
		},
		Reference: "CX23453XX",
		BenefitOf: "SomeName",
		AccountNo: "123-XX",
		Expires:   time.Now().Add(7 * time.Minute).Unix(),
	}

	header := http.Header{}
	header.Add("x-api-key", d.PartnerAPIKey)

	resTx := &NewTransactionResponse{}
	url := fmt.Sprintf("%s/company/%s/withdraw", d.QredoAPI, companyID)
	httpClient := util.NewHTTPClient()
	if err := httpClient.Request("POST", url, reqWD, resTx, header); err != nil {
		return nil, errors.Wrap(err, "request to partner api")
	}
	return resTx, nil
}

func (d *SaCli) ReadAction(feedUrl string) error {
	feed := d.Agent.ReadAction(feedUrl, nil)
	doneCH, stopCH, err := feed.ActionEvent(
		func(e *lib.WsActionInfoEvent) {
			fmt.Printf("action: id %s, status %v, type %v, agent-id %v\n", e.ID, e.Status, e.Type, e.AgentID)
		},
		func(err error) {
			fmt.Printf("err: %+v", err)
		},
	)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// handle signal interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		stopCH <- struct{}{}
		<-doneCH
		os.Exit(1)
	}()

	return nil
}

func (d *SaCli) sign(body []byte, timestamp, url string) (string, error) {
	if len(d.PrivateKey) == 0 {
		return "", errors.New("the private key was empty")
	}

	h := sha256.New()
	h.Write([]byte(timestamp))
	h.Write([]byte(url))
	h.Write(body)
	digest := h.Sum(nil)

	block, _ := pem.Decode([]byte(d.PrivateKey))
	rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", errors.Wrap(err, "parse RSA key")
	}

	signature, err := rsa.SignPKCS1v15(nil, rsaKey, crypto.SHA256, digest)
	if err != nil {
		return "", errors.Wrap(err, "sign request")
	}

	return base64.RawURLEncoding.EncodeToString(signature), nil
}
