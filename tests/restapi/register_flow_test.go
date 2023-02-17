package restapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/rest"
	"github.com/qredo/signing-agent/rest/version"
)

const (
	TestDataDBStoreFilePath               = "../../testdata/test-store.db"
	TestAPIKeyFilePath                    = "../../testdata/e2e/apikey"
	TestBase64PrivateKeyFilePath          = "../../testdata/e2e/base64privatekey"
	FixturePathRegisterClientInitResponse = "../../testdata/lib/registerClientInitResponse.json"
	TestBuildVersion                      = "(test-cb12berf)"
	TestBuildType                         = "test-dev"
	TestBuildDate                         = "Wed 29 Feb 2021 15:28:38 BST"
)

var (
	testAccountCode string
)

// Default creates configuration with default values
func testDefaultConf() config.Config {
	var cfg config.Config
	cfg.Default()
	cfg.Logging.Level = "debug"
	cfg.Store.FileConfig = TestDataDBStoreFilePath
	return cfg
}

func getTestLog() *zap.SugaredLogger {
	logConfig := zap.NewDevelopmentConfig()
	logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, _ := logConfig.Build()
	return l.Sugar()
}

func getTestHandlers(cfg config.Config) http.Handler {
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
	return router.SetHandlers()
}

func serverMockQB(QredoAPIBasePath string) *httptest.Server {
	handler := http.NewServeMux()
	uriClientInit := fmt.Sprintf("%s/coreclient/init", QredoAPIBasePath)
	handler.HandleFunc(uriClientInit, mockQBClientInit)

	uriClientInitFinish := fmt.Sprintf("%s/coreclient/finish", QredoAPIBasePath)
	handler.HandleFunc(uriClientInitFinish, mockQBClientInitFinish)

	uriClientFeed := fmt.Sprintf("%s/coreclient/feed", QredoAPIBasePath)
	handler.HandleFunc(uriClientFeed, mockQBClientFeed)

	uriActionFeed := fmt.Sprintf("%s/coreclient/action/{action_id}", QredoAPIBasePath)
	//uriActionFeed := fmt.Sprintf("%s/coreclient/action", QredoAPIBasePath)
	handler.HandleFunc(uriActionFeed, mockQBActionFeed)

	svr := httptest.NewServer(handler)
	return svr
}

func mockQBClientInit(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK

	fixtureFile, err := os.Open(FixturePathRegisterClientInitResponse)
	if err != nil {
		panic("Can't get fixtured ClientInit response file.")
	}

	dataFromFixture, err := io.ReadAll(fixtureFile)
	if err != nil {
		panic("Can't get fixtured ClientInit response content.")
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(dataFromFixture)
}

func mockQBClientInitFinish(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	response := &api.CoreClientServiceRegisterFinishResponse{
		Feed: fmt.Sprintf("ws://e2e-test-server/api/v1/p/coreclient/%s/feed", testAccountCode),
	}

	dataJSON, _ := json.Marshal(response)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(dataJSON)
}

func mockQBClientFeed(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	ws, _ := upgrader.Upgrade(w, r, nil)
	_ = ws.WriteControl(websocket.PingMessage, []byte{}, time.Time{})
	_ = ws.WriteMessage(websocket.TextMessage, []byte(`{"id": "Action if needed"}`))
	w.WriteHeader(http.StatusSwitchingProtocols)
}

// mockQBActionFeed mocks calls to backend to /coreclient/action/{action_id}.
func mockQBActionFeed(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		w.WriteHeader(http.StatusOK)
		var req api.CoreClientServiceActionApproveRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := &api.CoreClientServiceActionMessagesResponse{
		Messages: []string{"aa"},
	}

	dataJSON, _ := json.Marshal(response)
	w.Header().Add("content-type", "test/plain")
	_, _ = w.Write(dataJSON)
}

// TestRestAPIs  mocks the http server and tests each of the endpoints.
func TestRestAPIs(t *testing.T) {
	// API and Base64Private keys needed for the tests.  API key doesn't need to be real, but the Private
	// key is compared with data in FixturePathRegisterClientInitResponse.
	// defaults if not on the command line.
	APIKey := "not a real key"
	Base64PrivateKey, err := os.ReadFile(TestBase64PrivateKeyFilePath)
	assert.NoError(t, err)

	cfg := testDefaultConf()
	srvQB := serverMockQB("/api/v1/p")
	cfg.Base.QredoAPI = srvQB.URL + "/api/v1/p"
	cfg.AutoApprove = config.AutoApprove{
		Enabled:          true,
		RetryIntervalMax: 300,
		RetryInterval:    5,
	}
	cfg.Websocket = config.WebSocketConfig{
		ReconnectTimeOut:  200,
		ReconnectInterval: 15,
		QredoWebsocket:    "wss://play-api.qredo.network/api/v1/p/coreclient/feed",
		PingPeriod:        10,
	}
	handlers := getTestHandlers(cfg)
	server := httptest.NewServer(handlers)
	defer func() {
		if _, err := os.Stat(TestDataDBStoreFilePath); err == nil {
			err := os.Remove(TestDataDBStoreFilePath)
			assert.NoError(t, err)
		}
		server.Close()
	}()
	servAA := httpexpect.New(t, server.URL)

	// register agent
	payload := &api.ClientRegisterRequest{
		Name:             "Agent Test Name",
		APIKey:           string(APIKey),
		Base64PrivateKey: string(Base64PrivateKey),
	}

	// API endpoint tests
	healthcheckStatusTests(servAA, cfg.Websocket.QredoWebsocket, defs.ConnectionState.Closed, 0)
	registrationTests(servAA, payload)
	healthcheckVersionTests(servAA)
	healthCheckConfigTests(servAA)
	<-time.After(time.Second) //might take a second to open the websocket connection
	healthcheckStatusTests(servAA, cfg.Websocket.QredoWebsocket, defs.ConnectionState.Open, 0)
	clientActionTests(servAA)
	websocketTests(servAA, cfg.Websocket.QredoWebsocket)
}

// registrationTests checks the register endpoint (/register). The payload includes the data to be registered.
// Registration fails if the agent is already registered, otherwise the response include the new agentID
// and feed URL. Both conditions are tested.
func registrationTests(e *httpexpect.Expect, payload *api.ClientRegisterRequest) {

	// test registering a new agent
	registrationResponse := e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusOK).JSON()

	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("agentID").String().Equal("5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn")
	registrationResponse.Object().Value("feedURL").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")

	// GET: /client should return the same agentID
	testAccountCode = registrationResponse.Object().Value("agentID").Raw().(string)
	response := e.GET(rest.WrapPathPrefix(rest.PathClient)).
		Expect().
		Status(http.StatusOK)
	response.JSON().Object().NotEmpty()
	response.JSON().Object().Value("agentID").String().Equal(testAccountCode)
	response.JSON().Object().Value("feedURL").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")

	// Register an existing client should result in an error.
	registrationResponse = e.POST(rest.WrapPathPrefix(rest.PathClientFullRegister)).
		WithJSON(payload).
		Expect().
		Status(http.StatusBadRequest).JSON()
	registrationResponse.Object().NotEmpty()
	registrationResponse.Object().Value("Detail").String().Equal("AgentID already exist. You can not set new one.")
}

// healthCheckVersionTests checks the healthcheck version endpoint (/healthcheck/version).
func healthcheckVersionTests(e *httpexpect.Expect) {
	// GET: healthcheck/version
	hcVersion := e.GET(rest.WrapPathPrefix(rest.PathHealthcheckVersion)).
		Expect().
		Status(http.StatusOK).JSON()
	hcVersion.Object().NotEmpty()
	hcVersion.Object().Keys().ContainsOnly("buildVersion", "buildType", "buildDate")
	hcVersion.Object().ValueEqual("buildVersion", TestBuildVersion)
	hcVersion.Object().ValueEqual("buildType", TestBuildType)
	hcVersion.Object().ValueEqual("buildDate", TestBuildDate)
}

// healthCheckConfigTests checks the healthcheck config endpoint (/healthcheck/config).
func healthCheckConfigTests(e *httpexpect.Expect) {
	hcConfig := e.GET(rest.WrapPathPrefix(rest.PathHealthCheckConfig)).
		Expect().
		Status(http.StatusOK).JSON()

	hcConfig.Object().Keys().Contains("logging")
	logCfg := hcConfig.Object().Value("logging").Object()
	logCfg.Value("format").String().Equal("json")
	logCfg.Value("level").String().Equal("debug")

	hcConfig.Object().Keys().Contains("base")
	baseCfg := hcConfig.Object().Value("base").Object()
	baseCfg.Value("pin").Equal(0)
	baseCfg.Value("qredoAPI").NotNull()

	hcConfig.Object().Keys().Contains("autoApproval")
	autoApprove := hcConfig.Object().Value("autoApproval").Object()
	autoApprove.Value("enabled").Equal(true)
	autoApprove.Value("retryIntervalMaxSec").Equal(300)
	autoApprove.Value("retryIntervalSec").Equal(5)

	hcConfig.Object().Keys().Contains("websocket")
	websocket := hcConfig.Object().Value("websocket").Object()
	websocket.Value("reconnectTimeoutSec").Equal(200)
	websocket.Value("reconnectIntervalSec").Equal(15)

	hcConfig.Object().Keys().Contains("store")
	storeCfg := hcConfig.Object().Value("store").Object()
	storeCfg.Value("type").Equal("file")
	storeCfg.Value("file").Equal(TestDataDBStoreFilePath)

	hcConfig.Object().Keys().Contains("http")
	httpCfg := hcConfig.Object().Value("http").Object()
	httpCfg.Value("addr").String().Equal("127.0.0.1:8007")
	httpCfg.Value("CORSAllowOrigins").Array().Element(0).String().Equal("*")
	httpCfg.Value("logAllRequests").Equal(false)

	hcConfig.Object().Keys().Contains("loadBalancing")
	lbConfig := hcConfig.Object().Value("loadBalancing").Object()
	lbConfig.Value("enable").Equal(false)
	lbConfig.Value("onLockErrorTimeoutMs").Equal(300)
	lbConfig.Value("actionIDExpirationSec").Equal(6)

	lbConfig.Keys().Contains("redis")
	redisConfig := lbConfig.Value("redis").Object()
	redisConfig.Value("host").Equal("redis")
	redisConfig.Value("port").Equal(6379)
	redisConfig.Value("password").Equal("")
	redisConfig.Value("db").Equal(0)
}

// healthCheckStatusTests checks the healthcheck status endpoint (/healthcheck/status).
func healthcheckStatusTests(e *httpexpect.Expect, websocketUrl string, wsstatus string, connectedClients int) {
	hcStatus := e.GET(rest.WrapPathPrefix(rest.PathHealthCheckStatus)).
		Expect().
		Status(http.StatusOK).JSON()
	hcStatus.Object().NotEmpty()
	hcStatus.Object().Keys().ContainsOnly("websocket")
	webSocket := hcStatus.Object().Value("websocket").Object()
	webSocket.Value("readyState").String().Equal(wsstatus)
	webSocket.Value("remoteFeedURL").String().Equal(websocketUrl)
	webSocket.Value("localFeedURL").String().Equal("ws://127.0.0.1:8007/api/v1/client/feed")
	webSocket.Value("connectedClients").Equal(connectedClients)
}

// clientActionTest tests the action endpoints (/client/action/{action_id}.
func clientActionTests(e *httpexpect.Expect) {
	// action approve
	e.PUT(rest.WrapPathPrefix(rest.PathAction)).
		Expect().
		Status(http.StatusOK)

	// action reject
	e.DELETE(rest.WrapPathPrefix(rest.PathAction)).
		Expect().
		Status(http.StatusOK)
}

// websocketTest tests the signing agent websocket (/client/feed)
func websocketTests(e *httpexpect.Expect, host string) {
	ws := e.GET(rest.WrapPathPrefix(rest.PathClientFeed)).
		WithWebsocketUpgrade().
		Expect().
		Status(http.StatusSwitchingProtocols).
		Websocket()
	defer ws.Disconnect()

	//also check the number of connected clients is 1
	healthcheckStatusTests(e, host, defs.ConnectionState.Open, 1)

	ws.Subprotocol().Empty()
	ws.CloseWithText("bye", websocket.CloseNormalClosure)
}
