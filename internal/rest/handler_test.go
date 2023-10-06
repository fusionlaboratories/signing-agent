package rest

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
	"github.com/test-go/testify/assert"
)

type mockActionService struct {
	ApproveCalled bool
	RejectCalled  bool
	LastActionId  string
	NextError     error
}

func (m *mockActionService) Approve(actionID string) error {
	m.ApproveCalled = true
	m.LastActionId = actionID
	return m.NextError
}
func (m *mockActionService) Reject(actionID string) error {
	m.RejectCalled = true
	m.LastActionId = actionID
	return m.NextError
}

type mockAgentService struct {
	RegisterAgentCalled      bool
	RegisterClientFeedCalled bool
	StartCalled              bool
	GetAgentDetailsCalled    bool
	GetWebsocketStatusCalled bool

	NextError                     error
	NextStartError                error
	NextAgentRegisterResponse     *api.AgentRegisterResponse
	NextGetAgentDetailsResponse   *api.GetAgentDetailsResponse
	NextHealthCheckStatusResponse *api.HealthCheckStatusResponse

	LastRequest              *http.Request
	LastWriter               http.ResponseWriter
	LastAgentRegisterRequest *api.AgentRegisterRequest
}

func (m *mockAgentService) RegisterAgent(req *api.AgentRegisterRequest) (*api.AgentRegisterResponse, error) {
	m.RegisterAgentCalled = true
	m.LastAgentRegisterRequest = req
	return m.NextAgentRegisterResponse, m.NextError
}

func (m *mockAgentService) GetAgentDetails() (*api.GetAgentDetailsResponse, error) {
	m.GetAgentDetailsCalled = true
	return m.NextGetAgentDetailsResponse, m.NextError
}

func (m *mockAgentService) RegisterClientFeed(w http.ResponseWriter, r *http.Request) {
	m.RegisterClientFeedCalled = true
	m.LastRequest = r
	m.LastWriter = w
}

func (m *mockAgentService) Start() error {
	m.StartCalled = true
	return m.NextStartError
}

func (m *mockAgentService) Stop() {
}

func (m *mockAgentService) GetWebsocketStatus() *api.HealthCheckStatusResponse {
	m.GetWebsocketStatusCalled = true
	return m.NextHealthCheckStatusResponse
}

var testLog = util.NewTestLogger()

func NewTestRequest() *http.Request {
	test_req, _ := http.NewRequest("POST", "/path", bytes.NewReader([]byte(`
	{
		"APIKeyID":"test api key",
		"WorkspaceID":"test workspace",
		"Secret":"test secret"
	}`)))
	return test_req
}

func TestRouter_RegisterAgent_fails_to_decode_request(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{}

	var lastDecoded *http.Request
	decode := func(i interface{}, r *http.Request) error {
		lastDecoded = r
		return defs.ErrBadRequest().WithDetail("some decode error")
	}

	sut := &Router{
		agentService: agentSrvMock,
		log:          testLog,
		decode:       decode,
	}

	req, _ := http.NewRequest("POST", "/path", nil)

	//Act
	response, err := sut.RegisterAgent(nil, nil, req)

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "some decode error", detail)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.NotNil(t, lastDecoded)
	assert.Equal(t, lastDecoded, req)
	assert.False(t, agentSrvMock.RegisterAgentCalled)
}

func TestRouter_RegisterAgent_fails_to_register(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{
		NextError: fmt.Errorf("some error"),
	}

	sut := NewRouter(testLog, config.Config{}, api.Version{}, agentSrvMock, nil)

	//Act
	response, err := sut.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)
	assert.NotNil(t, err)
	assert.Equal(t, "some error", err.Error())
	assert.True(t, agentSrvMock.RegisterAgentCalled)

	assert.NotNil(t, agentSrvMock.LastAgentRegisterRequest)
	assert.Equal(t, "test api key", agentSrvMock.LastAgentRegisterRequest.APIKeyID)
	assert.Equal(t, "test workspace", agentSrvMock.LastAgentRegisterRequest.WorkspaceID)
	assert.Equal(t, "test secret", agentSrvMock.LastAgentRegisterRequest.Secret)
}

func TestRouter_RegisterAgent_fails_to_start_service(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{
		NextAgentRegisterResponse: &api.AgentRegisterResponse{},
		NextStartError:            fmt.Errorf("some error"),
	}

	sut := NewRouter(testLog, config.Config{}, api.Version{}, agentSrvMock, nil)

	//Act
	response, err := sut.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, response)

	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "failed to start the agent service. Please restart", detail)
	assert.Equal(t, http.StatusInternalServerError, code)
	assert.True(t, agentSrvMock.StartCalled)
}

func TestRouter_RegisterAgent_returns_resp(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{
		NextAgentRegisterResponse: &api.AgentRegisterResponse{
			GetAgentDetailsResponse: api.GetAgentDetailsResponse{
				Name:    "test name",
				AgentID: "test id",
				FeedURL: "test feed",
			},
		}}

	sut := NewRouter(testLog, config.Config{}, api.Version{}, agentSrvMock, nil)

	//Act
	response, err := sut.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, err)

	info, ok := response.(*api.AgentRegisterResponse)
	assert.True(t, ok)
	assert.Equal(t, "test id", info.AgentID)
	assert.Equal(t, "test feed", info.FeedURL)
	assert.Equal(t, "test name", info.Name)
}

func TestRouter_ClientFeed_registers_client(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{}
	handler := &Router{
		agentService: agentSrvMock,
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)
	w := httptest.NewRecorder()

	//Act
	res, err := handler.ClientFeed(nil, w, test_req)

	//Assert
	assert.Nil(t, res)
	assert.Nil(t, err)
	assert.True(t, agentSrvMock.RegisterClientFeedCalled)
	assert.Equal(t, test_req, agentSrvMock.LastRequest)
	assert.Equal(t, w, agentSrvMock.LastWriter)
}

func TestRouter_GetClient(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{
		NextGetAgentDetailsResponse: &api.GetAgentDetailsResponse{
			Name:    "test name",
			AgentID: "test id",
			FeedURL: "test feed",
		},
	}
	handler := &Router{
		agentService: agentSrvMock,
	}

	//Act
	response, err := handler.GetClient(nil, nil, nil)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, agentSrvMock.GetAgentDetailsCalled)

	info, ok := response.(*api.GetAgentDetailsResponse)
	assert.True(t, ok)
	assert.Equal(t, "test id", info.AgentID)
	assert.Equal(t, "test feed", info.FeedURL)
	assert.Equal(t, "test name", info.Name)
}

func TestRouter_ActionApprove_empty_actionId(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{}
	req, _ := http.NewRequest("PUT", "/client/action/ ", nil)
	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.False(t, actionSrvMock.ApproveCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	code, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRouter_ActionApprove_error_on_approve(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{
		NextError: errors.New("error while approving"),
	}
	req, _ := http.NewRequest("PUT", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Equal(t, "error while approving", err.Error())
	assert.True(t, actionSrvMock.ApproveCalled)
	assert.Nil(t, response)
}

func TestRouter_ActionApprove_success(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{}
	req, _ := http.NewRequest("PUT", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)

	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionApprove(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.NotNil(t, response)
	assert.True(t, actionSrvMock.ApproveCalled)
	assert.Nil(t, err)
	action_response, ok := response.(api.ActionResponse)
	assert.True(t, ok)
	assert.Equal(t, "some_action_id", action_response.ActionID)
	assert.Equal(t, "approved", action_response.Status)
}

func TestRouter_ActionReject_empty_actionId(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{}
	req, _ := http.NewRequest("DELETE", "/client/action/ ", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.False(t, actionSrvMock.RejectCalled)
	assert.NotNil(t, err)
	apiErr := err.(*defs.APIError)
	_, detail := apiErr.APIError()
	assert.Equal(t, "empty actionID", detail)
}

func TestRouter_ActionReject_error_on_reject(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{
		NextError: errors.New("error on reject"),
	}
	req, _ := http.NewRequest("DELETE", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)
	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.Nil(t, response)
	assert.True(t, actionSrvMock.RejectCalled)
	assert.Equal(t, "error on reject", err.Error())
}

func TestRouter_ActionReject_success(t *testing.T) {
	//Arrange
	actionSrvMock := &mockActionService{}
	req, _ := http.NewRequest("DELETE", "/client/action/some_action_id", nil)
	rr := httptest.NewRecorder()
	m := mux.NewRouter()
	var (
		err      error
		response interface{}
	)

	sut := NewRouter(testLog, config.Config{}, api.Version{}, nil, actionSrvMock)

	m.HandleFunc("/client/action/{action_id}", func(w http.ResponseWriter, r *http.Request) {
		response, err = sut.ActionReject(nil, w, r)
	})

	//Act
	m.ServeHTTP(rr, req)

	//Assert
	assert.NotNil(t, response)
	assert.True(t, actionSrvMock.RejectCalled)
	assert.Nil(t, err)
	action_response, ok := response.(api.ActionResponse)
	assert.True(t, ok)
	assert.Equal(t, "some_action_id", action_response.ActionID)
	assert.Equal(t, "rejected", action_response.Status)
}

func TestRouter_HealthCheckStatus(t *testing.T) {
	//Arrange
	agentSrvMock := &mockAgentService{
		NextHealthCheckStatusResponse: &api.HealthCheckStatusResponse{
			WebsocketStatus: api.WebsocketStatus{
				ReadyState:       "open",
				RemoteFeedUrl:    "remote test",
				ConnectedClients: 3,
			},
			LocalFeedUrl: "local test",
		},
	}

	sut := &Router{
		agentService: agentSrvMock,
	}

	//Act
	response, err := sut.HealthCheckStatus(nil, nil, nil)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)

	data, ok := response.(*api.HealthCheckStatusResponse)

	assert.True(t, ok)
	assert.Equal(t, "local test", data.LocalFeedUrl)
	assert.Equal(t, "open", data.WebsocketStatus.ReadyState)
	assert.Equal(t, "remote test", data.WebsocketStatus.RemoteFeedUrl)
	assert.Equal(t, uint32(3), data.WebsocketStatus.ConnectedClients)
}

func TestRouter_HealthCheckVersion(t *testing.T) {
	//Arrange
	version := &api.Version{
		BuildVersion: "some build version",
		BuildType:    "some build type",
		BuildDate:    "some build date",
	}
	sut := Router{
		version: *version,
	}

	//Act
	response, err := sut.HealthCheckVersion(nil, nil, nil)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)

	data, ok := response.(api.Version)
	assert.True(t, ok)
	assert.Equal(t, "some build date", data.BuildDate)
	assert.Equal(t, "some build version", data.BuildVersion)
	assert.Equal(t, "some build type", data.BuildType)
}

func TestRouter_HealthCheckConfig(t *testing.T) {
	//Arrange
	testConfig := config.Config{
		Base: config.Base{
			PIN:      25,
			QredoAPI: "some url",
		},
		HTTP: config.HttpSettings{
			Addr: "some address",
		},
	}
	sut := Router{
		config: testConfig,
	}

	//Act
	response, err := sut.HealthCheckConfig(nil, nil, nil)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)

	data, ok := response.(config.Config)
	assert.True(t, ok)
	assert.Equal(t, 25, data.Base.PIN)
	assert.Equal(t, "some url", data.Base.QredoAPI)
	assert.Equal(t, "some address", data.HTTP.Addr)
}
