package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/feed"
	"github.com/qredo/signing-agent/internal/hub"
	"github.com/qredo/signing-agent/internal/store"
	"github.com/qredo/signing-agent/internal/util"
	"github.com/test-go/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"
)

var (
	testLog          = util.NewTestLogger()
	ignoreOpenCensus = goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
)

type mockAutoApprover struct {
	StopCalled          bool
	ListenCalled        bool
	GetFeedClientCalled bool

	NextHubFeedClient *hub.HubFeedClient
}

type mockWebsocketUpgrader struct {
	UpgradeCalled           bool
	NextError               error
	NextWebsocketConnection hub.WebsocketConnection
	LastWriter              http.ResponseWriter
	LastRequest             *http.Request
	LastResponseHeader      http.Header
}

func (m *mockWebsocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (hub.WebsocketConnection, error) {
	m.UpgradeCalled = true
	m.LastWriter = w
	m.LastRequest = r
	m.LastResponseHeader = responseHeader
	return m.NextWebsocketConnection, m.NextError
}

type mockFeedHub struct {
	NextRun                  bool
	RunCalled                bool
	RegisterClientCalled     bool
	StopCalled               bool
	IsRunningCalled          bool
	GetWebsocketStatusCalled bool
	LastRegisteredClient     *hub.HubFeedClient

	NextWSstatus api.WebsocketStatus
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

func (m *mockFeedHub) RegisterClient(client *hub.HubFeedClient) {
	m.RegisterClientCalled = true
	m.LastRegisteredClient = client
}

func (m *mockFeedHub) UnregisterClient(client *hub.HubFeedClient) {
}

func (m *mockFeedHub) GetWebsocketStatus() api.WebsocketStatus {
	m.GetWebsocketStatusCalled = true
	return m.NextWSstatus
}

type mockStoreWriter struct {
	SaveAgentInfoCalled bool
	LastID              string
	LastAgent           *store.AgentInfo

	NextError error
}

func (m *mockStoreWriter) SaveAgentInfo(id string, agent *store.AgentInfo) error {
	m.SaveAgentInfoCalled = true
	m.LastID = id
	m.LastAgent = agent
	return m.NextError
}

type mockClientFeed struct {
	StartCalled         bool
	ListenCalled        bool
	GetFeedClientCalled bool
	NextFeedClient      *hub.HubFeedClient
}

func (m *mockClientFeed) Start(wg *sync.WaitGroup) {
	m.StartCalled = true
	//add a bit of delay to ensure the sync
	<-time.After(time.Second)
	wg.Done()
}

func (m *mockClientFeed) Listen(wg *sync.WaitGroup) {
	m.ListenCalled = true
	//add a bit of delay to ensure the sync
	<-time.After(2 * time.Second)
	wg.Done()
}

func (m *mockClientFeed) GetFeedClient() *hub.HubFeedClient {
	m.GetFeedClientCalled = true
	return m.NextFeedClient
}

func (m *mockAutoApprover) Listen(wg *sync.WaitGroup) {
	m.ListenCalled = true
	wg.Done()
}
func (m *mockAutoApprover) GetFeedClient() *hub.HubFeedClient {
	m.GetFeedClientCalled = true
	return m.NextHubFeedClient
}
func (m *mockAutoApprover) Stop() {
	m.StopCalled = true
}

func TestAgentService_Start_agent_not_registered_doesnt_run_hub(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	sut := NewAgentService(config.Config{}, nil, nil, nil, mockFeedHub, nil, util.NewTestLogger(),
		nil, nil)

	//Act
	err := sut.Start()

	//Assert
	assert.Nil(t, err)
	assert.False(t, mockFeedHub.RunCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestAgentService_Start_autoapprover_not_enabled_fail_to_run_hub(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: false,
	}
	sut := NewAgentService(config.Config{}, nil, nil, nil, mockFeedHub, nil, util.NewTestLogger(),
		nil, &store.AgentInfo{})

	//Act
	err := sut.Start()

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "failed to start the feed hub", err.Error())
	assert.True(t, mockFeedHub.RunCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestAgentService_Start_autoapprover_enabled_hub_doesnt_run(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: false,
	}

	mockApprover := &mockAutoApprover{}
	sut := agentSrv{
		log:          testLog,
		agentInfo:    &store.AgentInfo{},
		autoApprover: mockApprover,
		feedHub:      mockFeedHub,
	}

	//Act
	err := sut.Start()

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "failed to start the feed hub", err.Error())
	assert.True(t, mockFeedHub.RunCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
	assert.True(t, mockApprover.ListenCalled)
	assert.True(t, mockApprover.StopCalled)
}

func TestAgentService_Start_run_hub_autoapprove_not_enabled(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}

	sut := agentSrv{
		log:       testLog,
		agentInfo: &store.AgentInfo{},
		feedHub:   mockFeedHub,
	}
	//Act
	_ = sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestAgentService_Start_registers_auto_approval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockApprover := &mockAutoApprover{
		NextHubFeedClient: &hub.HubFeedClient{
			IsInternal: true,
		},
	}
	sut := agentSrv{
		log:          testLog,
		agentInfo:    &store.AgentInfo{},
		feedHub:      mockFeedHub,
		autoApprover: mockApprover,
	}

	//Act
	_ = sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockFeedHub.RegisterClientCalled)

	lastRegClient := mockFeedHub.LastRegisteredClient
	assert.NotNil(t, lastRegClient)
	assert.True(t, lastRegClient.IsInternal)
}

func TestAgentService_Stop_stops_feedhub_and_authProvider(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	authMock := &auth.MockHeaderProvider{}

	sut := NewAgentService(
		config.Config{}, authMock, nil, nil, mockFeedHub,
		nil, testLog, nil, nil)

	//Act
	sut.Stop()

	//Assert
	assert.True(t, mockFeedHub.StopCalled)
	assert.True(t, authMock.StopCalled)
}

func TestAgentService_RegisterClientFeed_hub_not_running(t *testing.T) {
	// Arrange
	mockFeedHub := &mockFeedHub{}
	sut := agentSrv{
		feedHub: mockFeedHub,
		log:     testLog,
	}

	// Act
	sut.RegisterClientFeed(nil, nil)

	// Assert
	assert.False(t, mockFeedHub.RegisterClientCalled)
	assert.True(t, mockFeedHub.IsRunningCalled)
}

func TestAgentService_RegisterClientFeed_upgrade_fails(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}

	mockUpgrader := &mockWebsocketUpgrader{
		NextError: errors.New("some upgrade error"),
	}
	sut := NewAgentService(config.Config{}, nil, nil, nil, mockFeedHub, nil, testLog, mockUpgrader, nil)

	test_req, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()

	//Act
	sut.RegisterClientFeed(w, test_req)

	//Assert
	assert.False(t, mockFeedHub.RegisterClientCalled)
	assert.True(t, mockUpgrader.UpgradeCalled)
	assert.Equal(t, test_req, mockUpgrader.LastRequest)
	assert.Equal(t, w, mockUpgrader.LastWriter)
	assert.Nil(t, mockUpgrader.LastResponseHeader)
}

func TestAgentService_RegisterClientFeed_client_registered(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}

	mockUpgrader := &mockWebsocketUpgrader{
		NextWebsocketConnection: &hub.MockWebsocketConnection{},
	}

	mockClientFeed := &mockClientFeed{
		NextFeedClient: &hub.HubFeedClient{},
	}

	sut := agentSrv{
		feedHub:  mockFeedHub,
		log:      testLog,
		upgrader: mockUpgrader,
		config:   config.Config{},
		newClientFeedFunc: func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister feed.UnregisterFunc, config config.WebSocketConfig) feed.ClientFeed {
			return mockClientFeed
		},
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()

	//Act
	sut.RegisterClientFeed(w, test_req)

	//Assert
	assert.True(t, mockFeedHub.RegisterClientCalled)
	assert.Equal(t, mockClientFeed.NextFeedClient, mockFeedHub.LastRegisteredClient)
	assert.True(t, mockUpgrader.UpgradeCalled)
	assert.Equal(t, test_req, mockUpgrader.LastRequest)
	assert.Equal(t, w, mockUpgrader.LastWriter)
	assert.Nil(t, mockUpgrader.LastResponseHeader)

	assert.True(t, mockClientFeed.StartCalled)
	assert.True(t, mockClientFeed.ListenCalled)
	assert.True(t, mockClientFeed.GetFeedClientCalled)
}

func TestAgentService_GetWebsocketStatus(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{
		NextWSstatus: api.WebsocketStatus{
			ReadyState:       "open",
			RemoteFeedUrl:    "some remote feed",
			ConnectedClients: 2,
		},
	}

	sut := agentSrv{
		feedHub: mockFeedHub,
		config: config.Config{
			HTTP: config.HttpSettings{
				Addr: "test-host",
			},
		},
	}

	//Act
	res := sut.GetWebsocketStatus()

	//Assert
	assert.True(t, mockFeedHub.GetWebsocketStatusCalled)
	assert.Equal(t, "ws://test-host/api/v2/client/feed", res.LocalFeedUrl)
	assert.Equal(t, "open", res.WebsocketStatus.ReadyState)
	assert.Equal(t, uint32(2), res.WebsocketStatus.ConnectedClients)
	assert.Equal(t, "some remote feed", res.WebsocketStatus.RemoteFeedUrl)
}

func TestAgentService_GetAgentDetails_agent_not_registered(t *testing.T) {
	//Arrange
	sut := agentSrv{
		log: testLog,
	}

	//Act
	res, err := sut.GetAgentDetails()

	//Assert
	assert.Nil(t, res)

	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "agent not registered", detail)
	assert.Equal(t, http.StatusNotFound, code)
}

func TestAgentService_GetAgentDetails_fails_to_get_agent_name(t *testing.T) {
	//Arrange
	testHeader := http.Header{}
	testHeader.Set("x-token", "test token")

	authMock := &auth.MockHeaderProvider{
		NextHeader: testHeader,
	}
	htcMock := util.NewHTTPMockClient()
	var (
		lastMethod      string
		lastURL         string
		lastHeaderValue string
	)

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		lastMethod = r.Method
		lastURL = r.URL.String()
		lastHeaderValue = r.Header.Get("x-token")

		return nil, errors.New("req error")

	}
	sut := agentSrv{
		log:          testLog,
		authProvider: authMock,
		htc:          htcMock,
		agentInfo: &store.AgentInfo{
			WorkspaceID: "wkspID",
			APIKeyID:    "apiKEyID",
		},
		config: config.Config{
			Base: config.Base{
				QredoAPI: "baseURL",
			},
		},
	}

	//Act
	res, err := sut.GetAgentDetails()

	//Assert
	assert.Nil(t, res)

	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to get agent name", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.Equal(t, http.MethodGet, lastMethod)
	assert.Equal(t, "baseURL/workspaces/wkspID/apikeys/apiKEyID", lastURL)
	assert.Equal(t, "test token", lastHeaderValue)
	assert.True(t, authMock.GetAuthHeaderCalled)
}

func TestAgentService_GetAgentDetails_returns_agent_info(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"name":"some name"}`))),
		}, nil
	}
	sut := agentSrv{
		log:          testLog,
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		agentInfo: &store.AgentInfo{
			APIKeyID: "apiKEyID",
		},
		config: config.Config{
			HTTP: config.HttpSettings{
				Addr: "some-address",
			},
		},
	}

	//Act
	res, err := sut.GetAgentDetails()

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, res)

	assert.Equal(t, "apiKEyID", res.AgentID)
	assert.Equal(t, "ws://some-address/api/v2/client/feed", res.FeedURL)
	assert.Equal(t, "some name", res.Name)
}

func TestAgentService_RegisterAgent_agent_already_registered(t *testing.T) {
	//Arrange
	sut := agentSrv{
		agentInfo: &store.AgentInfo{},
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "signing agent already registered", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestAgentService_RegisterAgent_fails_to_generate_keys(t *testing.T) {
	//Arrange
	sut := agentSrv{
		log: testLog,
		genKeysFunc: func() (string, string, string, string, error) {
			return "", "", "", "", errors.New("some gen error")
		},
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to generate keys", detail)
	assert.Equal(t, http.StatusInternalServerError, code)
}

func TestAgentService_RegisterAgent_fails_to_initiate_authProvider(t *testing.T) {
	//Arrange
	authMock := &auth.MockHeaderProvider{
		NextError: errors.New("some error"),
	}
	sut := agentSrv{
		authProvider: authMock,
		genKeysFunc:  defs.GenerateKeys,
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to initiate the auth provider", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.True(t, authMock.InitiateCalled)
	assert.Equal(t, "keyID", authMock.LastApiKeyID)
	assert.Equal(t, "secret", authMock.LastApiKeySecret)
	assert.Equal(t, "wkspID", authMock.LastWorkspaceID)
}

func TestAgentService_RegisterAgent_fails_to_updateAPIKey(t *testing.T) {
	//Arrange
	testHeader := http.Header{}
	testHeader.Set("x-token", "test token")

	authMock := &auth.MockHeaderProvider{
		NextHeader: testHeader,
	}

	htcMock := util.NewHTTPMockClient()
	var (
		lastMethod             string
		lastURL                string
		lastHeaderValue        string
		lastSaveKeyDataRequest saveKeyDataRequest
	)

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		lastMethod = r.Method
		lastURL = r.URL.String()
		lastHeaderValue = r.Header.Get("x-token")

		jd := json.NewDecoder(r.Body)

		defer r.Body.Close()
		_ = jd.Decode(&lastSaveKeyDataRequest)
		return nil, errors.New("req error")
	}

	sut := agentSrv{
		authProvider: authMock,
		htc:          htcMock,
		log:          util.NewTestLogger(),
		genKeysFunc: func() (string, string, string, string, error) {
			return "blsPubTest", "", "ecPubTest", "", nil
		},
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to register agent", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.Equal(t, http.MethodPut, lastMethod)
	assert.Equal(t, "/workspaces/wkspID/apikeys/keyID", lastURL)
	assert.Equal(t, "test token", lastHeaderValue)
	assert.True(t, authMock.GetAuthHeaderCalled)

	assert.Equal(t, "blsPubTest", lastSaveKeyDataRequest.BlsPublicKey)
	assert.Equal(t, "ecPubTest", lastSaveKeyDataRequest.EcPublicKey)
}

func TestAgentService_RegisterAgent_fails_to_setSignerKey(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"actionID":"testID"}`))),
		}, nil

	}

	signerMock := &action.MockSigner{
		NextSetKeyError: errors.New("some signer error"),
	}

	sut := agentSrv{
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		log:          util.NewTestLogger(),
		signer:       signerMock,
		genKeysFunc: func() (string, string, string, string, error) {
			return "", "blsPrivTest", "", "", nil
		},
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to setup signer", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.True(t, signerMock.SetKeyCalled)
	assert.Equal(t, "blsPrivTest", signerMock.LastBlsPrivateKey)
}

func TestAgentService_RegisterAgent_fails_to_approveAction(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"actionID":"testID"}`))),
		}, nil

	}

	signerMock := &action.MockSigner{
		NextError: errors.New("some approval error"),
	}

	sut := agentSrv{
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		log:          util.NewTestLogger(),
		signer:       signerMock,
		genKeysFunc:  defs.GenerateKeys,
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, "some approval error", err.Error())

	assert.True(t, signerMock.ActionApproveCalled)
	assert.Equal(t, "testID", signerMock.LastActionId)
}

func TestAgentService_RegisterAgent_fails_to_saveAgentInfo(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"actionID":"testID"}`))),
		}, nil

	}

	mockStore := &mockStoreWriter{
		NextError: errors.New("some db error"),
	}
	sut := agentSrv{
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		log:          util.NewTestLogger(),
		signer:       &action.MockSigner{},
		store:        mockStore,
		genKeysFunc: func() (string, string, string, string, error) {
			return "", "blsPrivTest", "", "ecPrivTest", nil
		},
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to save agent info", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.True(t, mockStore.SaveAgentInfoCalled)
	assert.Equal(t, "keyID", mockStore.LastID)

	assert.Equal(t, "keyID", mockStore.LastAgent.APIKeyID)
	assert.Equal(t, "wkspID", mockStore.LastAgent.WorkspaceID)
	assert.Equal(t, "secret", mockStore.LastAgent.APIKeySecret)
	assert.Equal(t, "blsPrivTest", mockStore.LastAgent.BLSPrivateKey)
	assert.Equal(t, "ecPrivTest", mockStore.LastAgent.ECPrivateKey)
	assert.NotEmpty(t, mockStore.LastAgent.BLSPrivateKey)
	assert.NotEmpty(t, mockStore.LastAgent.ECPrivateKey)
}

func TestAgentService_RegisterAgent_fails_to_get_agentName(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	isGetName := false
	var (
		lastMethod string
		lastURL    string
	)
	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		if !isGetName {
			isGetName = true
			return &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"actionID":"testID"}`))),
			}, nil
		}
		lastMethod = r.Method
		lastURL = r.URL.String()

		return nil, errors.New("some get name error")
	}

	sut := agentSrv{
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		log:          util.NewTestLogger(),
		signer:       &action.MockSigner{},
		store:        &mockStoreWriter{},
		config: config.Config{
			HTTP: config.HttpSettings{
				Addr: "localaddress",
			},
		},
		genKeysFunc: defs.GenerateKeys,
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, res)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to get agent name", detail)
	assert.Equal(t, http.StatusInternalServerError, code)

	assert.Equal(t, http.MethodGet, lastMethod)
	assert.Equal(t, "/workspaces/wkspID/apikeys/keyID", lastURL)
}

func TestAgentService_RegisterAgent_returns_agentInfo(t *testing.T) {
	//Arrange
	htcMock := util.NewHTTPMockClient()
	isGetName := false

	util.GetDoMockHTTPClientFunc = func(r *http.Request) (*http.Response, error) {
		if !isGetName {
			isGetName = true
			return &http.Response{
				Status:     "200 OK",
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"actionID":"testID"}`))),
			}, nil
		}

		return &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"name":"someName"}`))),
		}, nil
	}

	sut := agentSrv{
		authProvider: &auth.MockHeaderProvider{},
		htc:          htcMock,
		log:          util.NewTestLogger(),
		signer:       &action.MockSigner{},
		store:        &mockStoreWriter{},
		config: config.Config{
			HTTP: config.HttpSettings{
				Addr: "localaddress",
			},
		},
		genKeysFunc: defs.GenerateKeys,
	}

	//Act
	res, err := sut.RegisterAgent(&api.AgentRegisterRequest{
		APIKeyID:    "keyID",
		Secret:      "secret",
		WorkspaceID: "wkspID",
	})

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "keyID", res.AgentID)
	assert.Equal(t, "ws://localaddress/api/v2/client/feed", res.FeedURL)
	assert.Equal(t, "someName", res.Name)
}
