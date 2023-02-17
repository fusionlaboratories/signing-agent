package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/autoapprover"
	"github.com/qredo/signing-agent/clientfeed"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

type mockFeedHub struct {
	NextRun                bool
	RunCalled              bool
	RegisterClientCalled   bool
	UnregisterClientCalled bool
	StopCalled             bool
	IsRunningCalled        bool
	LastRegisteredClient   *hub.FeedClient
	LastUnregisteredClient *hub.FeedClient
}

func (m *mockFeedHub) IsRunning() bool {
	m.IsRunningCalled = true
	return m.NextRun
}

func (m *mockFeedHub) Run() bool {
	m.RunCalled = true
	return m.NextRun
}

func (m *mockFeedHub) Stop() {
	m.StopCalled = true
}

func (m *mockFeedHub) RegisterClient(client *hub.FeedClient) {
	m.RegisterClientCalled = true
	m.LastRegisteredClient = client
}

func (m *mockFeedHub) UnregisterClient(client *hub.FeedClient) {
	m.UnregisterClientCalled = true
	m.LastUnregisteredClient = client
}

func (m *mockFeedHub) GetExternalFeedClients() int {
	return 0
}

type mockClientFeed struct {
	StartCalled         bool
	ListenCalled        bool
	GetFeedClientCalled bool
	NextFeedClient      *hub.FeedClient
}

func (m *mockClientFeed) Start(wg *sync.WaitGroup) {
	m.StartCalled = true
	wg.Done()
}

func (m *mockClientFeed) Listen() {
	m.ListenCalled = true
}
func (m *mockClientFeed) GetFeedClient() *hub.FeedClient {
	m.GetFeedClientCalled = true
	return m.NextFeedClient
}

var testLog = util.NewTestLogger()

func NewTestRequest() *http.Request {
	test_req, _ := http.NewRequest("POST", "/path", bytes.NewReader([]byte(`
	{
		"Name":"test name",
		"APIKey":"test api key",
		"Base64PrivateKey":"test 64 private key"
	}`)))
	return test_req
}

var testClientRegisterResponse = &api.ClientRegisterResponse{
	ECPublicKey:  "ec",
	BLSPublicKey: "bls",
	RefID:        "refId",
}

var testRegisterInitResponse = &api.QredoRegisterInitResponse{
	ID:           "some id",
	ClientID:     "client id",
	ClientSecret: "client secret",
	AccountCode:  "account code",
	IDDocument:   "iddocument",
	Timestamp:    15456465,
}

func TestSigningAgentHandler_RegisterAgent_already_registered(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextAgentID: "some agent id",
	}
	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil, nil, "")

	rr := httptest.NewRecorder()

	//Act
	response, err := handler.RegisterAgent(nil, rr, nil)

	//Assert
	assert.Nil(t, response)
	assert.True(t, mock_core.GetSystemAgentIDCalled)
	assert.NotNil(t, err)

	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "AgentID already exist. You can not set new one.", detail)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.False(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_decode_request(t *testing.T) {
	//Arrange
	mock_core := lib.NewMockSigningAgentClient("")

	var lastDecoded *http.Request
	decode := func(i interface{}, r *http.Request) error {
		lastDecoded = r
		return defs.ErrBadRequest().WithDetail("some decode error")
	}

	handler := &SigningAgentHandler{
		feedHub: &mockFeedHub{},
		core:    mock_core,
		log:     testLog,
		decode:  decode,
	}

	req, _ := http.NewRequest("POST", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.RegisterAgent(nil, rr, req)

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "some decode error", detail)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.True(t, mock_core.GetSystemAgentIDCalled)
	assert.False(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
	assert.NotNil(t, lastDecoded)
	assert.Equal(t, lastDecoded, req)
}

func TestSigningAgentHandler_RegisterAgent_doesnt_validate_request(t *testing.T) {
	//Arrange
	mock_core := lib.NewMockSigningAgentClient("")

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil, nil, "")

	req, _ := http.NewRequest("POST", "/path", bytes.NewReader([]byte(`
	{
		"APIKey":"key",
		"Base64PrivateKey":"key"
	}`)))

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), req)

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "name", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_register_client(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextRegisterError: errors.New("some error"),
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil, nil, "")

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
	assert.True(t, mock_core.ClientRegisterCalled)
	assert.False(t, mock_core.ClientInitCalled)
	assert.Equal(t, "test name", mock_core.LastRegisteredName)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_init_registration(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextClientInitError:        errors.New("some error"),
		NextClientRegisterResponse: testClientRegisterResponse,
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil, nil, "")

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.True(t, mock_core.ClientInitCalled)
	assert.Equal(t, "some error", err.Error())
	assert.Equal(t, "refId", mock_core.LastRef)
	assert.Equal(t, "test api key", mock_core.LastApiKey)
	assert.Equal(t, "test 64 private key", mock_core.Last64PrivateKey)
	assert.Equal(t, "bls", mock_core.LastRegisterRequest.BLSPublicKey)
	assert.Equal(t, "ec", mock_core.LastRegisterRequest.ECPublicKey)
	assert.Equal(t, "test name", mock_core.LastRegisterRequest.Name)
}

func TestSigningAgentHandler_RegisterAgent_fails_to_finish_registration(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextRegisterFinishError:    errors.New("some error"),
		NextClientRegisterResponse: testClientRegisterResponse,
		NextRegisterInitResponse:   testRegisterInitResponse,
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{}}, nil, nil, "")

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
	assert.True(t, mock_core.ClientRegisterFinishCalled)
	assert.Equal(t, "refId", mock_core.LastRef)

	assert.Equal(t, "account code", mock_core.LastRegisterFinishRequest.AccountCode)
	assert.Equal(t, "client id", mock_core.LastRegisterFinishRequest.ClientID)
	assert.Equal(t, "client secret", mock_core.LastRegisterFinishRequest.ClientSecret)
	assert.Equal(t, "some id", mock_core.LastRegisterFinishRequest.ID)
	assert.Equal(t, "iddocument", mock_core.LastRegisterFinishRequest.IDDocument)
}

func TestSigningAgentHandler_RegisterAgent_returns_response(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextClientRegisterResponse: testClientRegisterResponse,
		NextRegisterInitResponse:   testRegisterInitResponse,
		NextRegisterFinishResponse: &api.ClientRegisterFinishResponse{},
	}

	handler := NewSigningAgentHandler(&mockFeedHub{}, mock_core, testLog, &config.Config{
		HTTP: config.HttpSettings{
			Addr: "some address",
		}}, nil, nil, "ws://some address/api/v1/client/feed")

	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, mock_core.ClientRegisterFinishCalled)

	res, ok := response.(api.AgentRegisterResponse)
	assert.True(t, ok)
	assert.NotNil(t, res)
	assert.Equal(t, "account code", res.AgentID)
	assert.Equal(t, "ws://some address/api/v1/client/feed", res.FeedURL)
}

func TestSigningAgentHandler_StartAgent_runs_feedHub(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")
	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog,
		&config.Config{
			HTTP: config.HttpSettings{}}, nil, nil, "")

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestSigningAgentHandler_StartAgent_doesnt_start_autoApproval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")
	config := &config.Config{
		HTTP:        config.HttpSettings{},
		AutoApprove: config.AutoApprove{},
	}
	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog, config, nil, nil, "")

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestSigningAgentHandler_StartAgent_registers_auto_approval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")

	handler := NewSigningAgentHandler(mockFeedHub, mockCore, testLog, &config.Config{
		AutoApprove: config.AutoApprove{
			Enabled: true,
		},
	}, autoapprover.NewAutoApprover(mockCore, testLog, &config.Config{}, nil), nil, "")

	//Act
	handler.StartAgent()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.True(t, mockFeedHub.RegisterClientCalled)

	auto_approver := mockFeedHub.LastRegisteredClient
	assert.NotNil(t, auto_approver)
	assert.True(t, auto_approver.IsInternal)
	close(auto_approver.Feed)
}

func TestSigningAgentHandler_StopAgent(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	handler := NewSigningAgentHandler(mockFeedHub, nil, util.NewTestLogger(), &config.Config{
		HTTP: config.HttpSettings{}}, &autoapprover.AutoApprover{}, nil, "")

	//Act
	handler.StopAgent()

	//Assert
	assert.True(t, mockFeedHub.StopCalled)
}

func TestSigningAgentHandler_ClientFeed_hub_not_running(t *testing.T) {
	//Arrange
	mockHub := &mockFeedHub{}
	handler := &SigningAgentHandler{
		feedHub: mockHub,
		log:     testLog,
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)

	//Act
	response, err := handler.ClientFeed(nil, httptest.NewRecorder(), test_req)

	//Assert
	assert.Nil(t, response)
	assert.Nil(t, err)
	assert.False(t, mockHub.RegisterClientCalled)
}

func TestSigningAgentHandler_ClientFeed_upgrade_fails(t *testing.T) {
	//Arrange
	mockHub := &mockFeedHub{
		NextRun: true,
	}
	mockUpgrader := &hub.MockWebsocketUpgrader{
		NextError: errors.New("some error"),
	}
	handler := &SigningAgentHandler{
		feedHub:  mockHub,
		log:      testLog,
		upgrader: mockUpgrader,
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()

	//Act
	response, err := handler.ClientFeed(nil, w, test_req)

	//Assert
	assert.Nil(t, response)
	assert.Nil(t, err)
	assert.False(t, mockHub.RegisterClientCalled)
	assert.True(t, mockUpgrader.UpgradeCalled)
	assert.Equal(t, test_req, mockUpgrader.LastRequest)
	assert.Equal(t, w, mockUpgrader.LastWriter)
	assert.Nil(t, mockUpgrader.LastResponseHeader)
}

func TestSigningAgentHandler_ClientFeed_registers_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockHub := &mockFeedHub{
		NextRun: true,
	}
	mockConn := &hub.MockWebsocketConnection{}
	mockUpgrader := &hub.MockWebsocketUpgrader{
		NextWebsocketConnection: mockConn,
	}
	feedClient := hub.NewFeedClient(true)
	mockFeedClient := &mockClientFeed{
		NextFeedClient: &feedClient,
	}
	newClientfunc := func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister clientfeed.UnregisterFunc, config *config.WebSocketConfig) clientfeed.ClientFeed {
		return mockFeedClient
	}
	handler := &SigningAgentHandler{
		feedHub:           mockHub,
		log:               testLog,
		upgrader:          mockUpgrader,
		websocketConfig:   &config.WebSocketConfig{},
		newClientFeedFunc: newClientfunc,
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)

	//Act
	_, _ = handler.ClientFeed(nil, httptest.NewRecorder(), test_req)
	<-time.After(time.Second)

	//Assert
	assert.True(t, mockHub.RegisterClientCalled)
	assert.NotNil(t, mockHub.LastRegisteredClient)
	assert.Equal(t, &feedClient, mockHub.LastRegisteredClient)

	assert.True(t, mockFeedClient.StartCalled)
	assert.True(t, mockFeedClient.GetFeedClientCalled)
	assert.True(t, mockFeedClient.ListenCalled)
}

func TestSigningAgentHandler_GetClient(t *testing.T) {
	//Arrange
	mockCore := &lib.MockSigningAgentClient{
		NextAgentID: "client 1",
	}
	handler := &SigningAgentHandler{
		core:      mockCore,
		localFeed: "feed/path",
	}
	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.GetClient(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)

	assert.True(t, mockCore.GetAgentIDCalled)
	data, _ := json.Marshal(response)

	assert.Equal(t, "{\"agentID\":\"client 1\",\"agentName\":\"\",\"feedURL\":\"feed/path\"}", string(data))
}
