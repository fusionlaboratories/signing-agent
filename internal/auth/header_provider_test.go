package auth

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/internal/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var ignoreOpenCensus = goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")

func TestHeaderProvider_Initiate_fails_to_initToken(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	var (
		lastMethod string
		lastURL    string
		lastHeader http.Header
	)

	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		lastMethod = r.Method
		lastURL = r.URL.String()
		lastHeader = r.Header
		return nil, errors.New("some req error")
	}

	sut := NewHeaderProvider("baseURL", util.NewTestLogger())
	sut.(*apiTokenProvider).htc = htcMock

	//Act
	err := sut.Initiate("wkspID", "test secret", "test key id")

	//Assert
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error while getting token response")
	assert.Equal(t, http.MethodGet, lastMethod)
	assert.Equal(t, "baseURL/workspaces/wkspID/token", lastURL)

	assert.Equal(t, "test key id", lastHeader.Get("Qredo-api-key"))
	assert.NotEmpty(t, "test key id", lastHeader.Get("Qredo-api-timestamp"))
	assert.NotEmpty(t, "test key id", lastHeader.Get("Qredo-api-signature"))
}

func TestHeaderProvider_Initiate_setsToken(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	htcMock := util.NewHTTPMockClient()

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte("{\"token\":\"testToken\"}"))),
		}, nil

	}

	var lastTokenValue string

	sut := apiTokenProvider{
		baseURL: "baseURL",
		htc:     htcMock,
		getTokenDurationFunc: func(value string) time.Duration {
			lastTokenValue = value
			return 5 * time.Minute
		},
		log:  util.NewTestLogger(),
		stop: make(chan bool),
		lock: sync.RWMutex{},
	}

	//Act
	err := sut.Initiate("wkspID", "test secret", "test key id")

	//Assert
	assert.Nil(t, err)
	assert.Equal(t, 5*time.Minute, sut.tokenTTL)
	assert.Equal(t, "testToken", lastTokenValue)
	sut.Stop()
}

func TestHeaderProvider_refreshToken(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	lastURLs := make(map[string]bool)
	htcMock := util.NewHTTPMockClient()

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		lastURLs[r.URL.String()] = true

		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte("{\"token\":\"testToken\"}"))),
		}, nil
	}

	sut := apiTokenProvider{
		baseURL: "baseURL",
		htc:     htcMock,
		getTokenDurationFunc: func(value string) time.Duration {
			return time.Second
		},
		log:  util.NewTestLogger(),
		stop: make(chan bool),
		lock: sync.RWMutex{},
	}

	err := sut.Initiate("wkspID", "test secret", "test key id")
	<-time.After(4 * time.Second)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(lastURLs))

	assert.True(t, lastURLs["baseURL/workspaces/wkspID/token"])
	assert.True(t, lastURLs["baseURL/workspaces/wkspID/token/refresh"])
	sut.Stop()
}
