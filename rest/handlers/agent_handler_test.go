package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

type mockManager struct {
	RegisterClientFeedCalled bool
	StartCalled              bool
	LastRequest              *http.Request
}

func (m *mockManager) RegisterClientFeed(_ http.ResponseWriter, r *http.Request) {
	m.RegisterClientFeedCalled = true
	m.LastRequest = r
}

func (m *mockManager) Start() {
	m.StartCalled = true
}

func (m *mockManager) Stop() {
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
	handler := NewSigningAgentHandler(nil, mock_core, testLog, "")

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
		core:   mock_core,
		log:    testLog,
		decode: decode,
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

	handler := NewSigningAgentHandler(nil, mock_core, testLog, "")

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

	handler := NewSigningAgentHandler(nil, mock_core, testLog, "")

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

	handler := NewSigningAgentHandler(nil, mock_core, testLog, "")

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

	handler := NewSigningAgentHandler(nil, mock_core, testLog, "")

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

func TestSigningAgentHandler_RegisterAgent_starts_manager(t *testing.T) {
	//Arrange
	mock_core := &lib.MockSigningAgentClient{
		NextClientRegisterResponse: testClientRegisterResponse,
		NextRegisterInitResponse:   testRegisterInitResponse,
		NextRegisterFinishResponse: &api.ClientRegisterFinishResponse{},
	}

	mockManager := &mockManager{}
	handler := NewSigningAgentHandler(mockManager, mock_core, testLog, "ws://some address/api/v1/client/feed")
	//Act
	response, err := handler.RegisterAgent(nil, httptest.NewRecorder(), NewTestRequest())

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.True(t, mock_core.ClientRegisterFinishCalled)

	assert.True(t, mockManager.StartCalled)

	res, ok := response.(api.AgentRegisterResponse)
	assert.True(t, ok)
	assert.NotNil(t, res)
	assert.Equal(t, "account code", res.AgentID)
	assert.Equal(t, "ws://some address/api/v1/client/feed", res.FeedURL)
}

func TestSigningAgentHandler_ClientFeed_registers_client(t *testing.T) {
	//Arrange
	mockManager := &mockManager{}
	handler := &SigningAgentHandler{
		agentManager: mockManager,
	}

	test_req, _ := http.NewRequest("GET", "/path", nil)

	//Act
	res, err := handler.ClientFeed(nil, httptest.NewRecorder(), test_req)

	//Assert
	assert.Nil(t, res)
	assert.Nil(t, err)

	assert.True(t, mockManager.RegisterClientFeedCalled)
	assert.Equal(t, test_req, mockManager.LastRequest)
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
