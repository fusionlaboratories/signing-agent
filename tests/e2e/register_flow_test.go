package e2e_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/rest"
	"github.com/qredo/signing-agent/rest/version"
)

const (
	TestQredoAPI            = "https://play-api.qredo.network/api/v1/p"
	TestDataDBStoreFilePath = "../../testdata/test-store.db"
	TestBuildVersion        = "(test-cb12berf)"
	TestBuildType           = "test-dev"
	TestBuildDate           = "Wed 29 Feb 2021 15:28:38 BST"
)

var (
	APIKey           string
	Base64PrivateKey string
)

// Default creates configuration with default values
func createTestConfig() config.Config {
	// can't proceed without an APIKEY or BASE64PKEY
	APIKey = os.Getenv("APIKEY")
	if APIKey == "" {
		log.Fatalf("APIKEY not set in environment")
	}
	Base64PrivateKey = os.Getenv("BASE64PKEY")
	if Base64PrivateKey == "" {
		log.Fatalf("BASE64PKEY not set in environment")
	}

	var cfg config.Config
	cfg.Default()
	cfg.Logging.Level = "debug"
	cfg.Store.FileConfig = TestDataDBStoreFilePath
	cfg.Websocket.QredoWebsocket = "wss://play-api.qredo.network/api/v1/p/coreclient/feed"
	cfg.AutoApprove.Enabled = true
	cfg.Base.QredoAPI = TestQredoAPI
	return cfg
}

func getTestLog() *zap.SugaredLogger {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, _ := logConfig.Build()
	return l.Sugar()
}

func getTestServer(cfg config.Config) *httptest.Server {
	//func createHttpServer(cfg config.Config) http.Server {
	log := getTestLog()
	ver := version.DefaultVersion()
	ver.BuildVersion = TestBuildVersion
	ver.BuildType = TestBuildType
	ver.BuildDate = TestBuildDate

	router, err := rest.NewQRouter(log, &cfg, ver)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	return httptest.NewServer(router.SetHandlers())
}

// TestRegisterNewSigningAgent tests initialising a new agent and checks if the agentID can be looked up.
func TestRegisterNewSigningAgent(t *testing.T) {

	cfg := createTestConfig()
	server := getTestServer(cfg)

	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()

	e := httpexpect.New(t, server.URL)

	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// register the client and check against response against expected
	register_response := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).WithJSON(payload).Expect()
	registrationResponse := register_response.Status(http.StatusOK).JSON()

	register_response.Header("Content-Type").Equal("application/json")
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("agentID").String().NotEmpty()
	registrationResponse.Object().Value("feedURL").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")
	agentID := registrationResponse.Object().Value("agentID").Raw().(string)

	// GET: /client should return the same agentID
	response := e.GET(rest.WrapPathPrefix(rest.PathClient)).
		Expect().
		Status(http.StatusOK)
	response.JSON().Object().NotEmpty()
	response.JSON().Object().Value("agentID").String().NotEmpty()
	response.JSON().Object().Value("agentID").String().Equal(agentID)
	response.JSON().Object().Value("feedURL").String().NotEmpty()
	response.JSON().Object().Value("feedURL").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")
	response.Header("Content-Type").Equal("application/json")
}

// TestRegisterExistingAgentDeny checks attempting to register fails with a known message.
func TestRegisterExistingAgentDeny(t *testing.T) {
	cfg := createTestConfig()
	server := getTestServer(cfg)

	defer func() {
		err := os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
		server.Close()
	}()

	e := httpexpect.New(t, server.URL)

	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// a new client
	e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	// registering again should result in an error
	response := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).WithJSON(payload).Expect()
	registrationResponse := response.Status(http.StatusBadRequest).JSON()
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("Detail").String().Equal("AgentID already exist. You can not set new one.")
	response.Header("Content-Type").Equal("application/json")
}
