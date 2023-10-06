package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
	"go.uber.org/zap"
)

const (
	apiKeyAuthHeader   = "qredo-api-key"
	apiSignatureHeader = "qredo-api-signature"
	apiTimestampHeader = "qredo-api-timestamp"
	authHeader         = "x-token"
)

type getTokenResponse struct {
	Token string `json:"token"`
}

type HeaderProvider interface {
	Initiate(workspaceID, apiKeySecret, apiKeyID string) error
	GetAuthHeader() http.Header
	Stop()
}

type apiTokenProvider struct {
	baseURL      string
	workspaceID  string
	apiKeySecret []byte
	apiKeyID     string
	token        string

	htc                  *util.Client
	tokenTTL             time.Duration
	lock                 sync.RWMutex
	log                  *zap.SugaredLogger
	stop                 chan bool
	getTokenDurationFunc func(string) time.Duration
}

func NewHeaderProvider(baseURL string, log *zap.SugaredLogger) HeaderProvider {
	return &apiTokenProvider{
		htc:                  util.NewHTTPClient(),
		baseURL:              baseURL,
		log:                  log,
		stop:                 make(chan bool),
		lock:                 sync.RWMutex{},
		getTokenDurationFunc: getTokenDuration,
	}
}

func (p *apiTokenProvider) Initiate(workspaceID, apiKeySecret, apiKeyID string) error {
	p.apiKeySecret, _ = base64.RawURLEncoding.DecodeString(apiKeySecret)
	p.apiKeyID = apiKeyID
	p.workspaceID = workspaceID

	if err := p.initToken(); err != nil {
		p.log.Errorf("failed to initialize token, err: %v", err)
		return err
	}

	// set a ticker to half the token's validity time
	ticker := time.NewTicker(p.tokenTTL / 2)

	go func() {
		defer func() {
			ticker.Stop()
			p.log.Info("HeaderProvider: stopped")
		}()

		for {
			select {
			case <-ticker.C:
				p.log.Info("HeaderProvider: token expired, refreshing")

				if !p.refreshToken() {
					// failed to refresh the existing token, issue a new token
					if err := p.initToken(); err != nil {
						p.log.Errorf("failed to initialize token, err: %v", err)
						return
					}
				}

				ticker.Reset(p.tokenTTL / 2)

			case <-p.stop:
				return
			}
		}
	}()

	return nil
}

func (p *apiTokenProvider) Stop() {
	close(p.stop)
}

func (p *apiTokenProvider) GetAuthHeader() http.Header {
	p.lock.RLock()
	defer p.lock.RUnlock()

	header := http.Header{}
	header.Set(authHeader, p.token)
	return header
}

func (p *apiTokenProvider) initToken() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.log.Info("HeaderProvider: initiating token")

	url := defs.URLToken(p.baseURL, p.workspaceID)
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	method := http.MethodGet

	body := []byte{}
	sig, err := defs.HmacSum(timestamp, method, url, p.apiKeySecret, body)
	if err != nil {
		return fmt.Errorf("failed to generate signature, HmacSum err: %v", err)
	}

	sigb64 := base64.RawURLEncoding.EncodeToString(sig)

	header := http.Header{}
	header.Add(apiTimestampHeader, timestamp)
	header.Add(apiKeyAuthHeader, p.apiKeyID)
	header.Add(apiSignatureHeader, sigb64)

	resp := &getTokenResponse{}
	if err := p.htc.Request(method, url, nil, resp, header); err != nil {
		return fmt.Errorf("error while getting token response, err: %v", err)
	}

	p.token = resp.Token
	p.tokenTTL = p.getTokenDurationFunc(p.token)

	if p.tokenTTL == 0 {
		return fmt.Errorf("invalid token duration")
	}

	p.log.Debugf("HeaderProvider: token validity %v", p.tokenTTL/2)
	return nil
}

func (p *apiTokenProvider) refreshToken() bool {
	url := defs.URLTokenRefresh(p.baseURL, p.workspaceID)
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	p.lock.RLock()
	req.Header.Add(authHeader, p.token)
	p.lock.RUnlock()

	resp := &getTokenResponse{}
	if err := p.htc.Request(http.MethodGet, url, nil, resp, p.GetAuthHeader()); err != nil {
		p.log.Errorf("HeaderProvider: error while refreshing token, err: %v", err)
		return false
	}

	if resp.Token == defs.EmptyString {
		p.log.Error("HeaderProvider: empty token response")
		return false
	}

	p.lock.Lock()
	p.token = resp.Token
	p.lock.Unlock()

	p.tokenTTL = p.getTokenDurationFunc(resp.Token)
	return true
}

func getTokenDuration(value string) time.Duration {
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return 0
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0
	}

	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return 0
	}

	expiration, ok := claims["exp"].(float64)
	if !ok {
		return 0
	}

	return time.Until(time.Unix(int64(expiration), 0))
}
