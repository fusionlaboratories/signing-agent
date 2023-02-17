package lib

import "github.com/qredo/signing-agent/api"

type MockSigningAgentClient struct {
	GetSystemAgentIDCalled     bool
	GetAgentZKPOnePassCalled   bool
	ClientRegisterCalled       bool
	ClientInitCalled           bool
	ClientRegisterFinishCalled bool
	ActionApproveCalled        bool
	GetAgentIDCalled           bool
	GetAgentNameCalled         bool
	ActionRejectCalled         bool
	Counter                    int
	NextError                  error
	NextClientInitError        error
	NextRegisterError          error
	NextRegisterFinishError    error
	NextZKPOnePass             []byte
	NextAgentID                string
	NextAgentName              string
	LastRegisteredName         string
	NextClientRegisterResponse *api.ClientRegisterResponse
	NextRegisterInitResponse   *api.QredoRegisterInitResponse
	NextRegisterFinishResponse *api.ClientRegisterFinishResponse
	LastRegisterRequest        *api.QredoRegisterInitRequest
	LastRegisterFinishRequest  *api.ClientRegisterFinishRequest
	LastRef                    string
	LastApiKey                 string
	Last64PrivateKey           string
	LastActionId               string
	LastRejectActionId         string
	LastAgentID                string
}

func NewMockSigningAgentClient(agentId string) *MockSigningAgentClient {
	return &MockSigningAgentClient{
		NextAgentID: agentId,
	}
}

func (m *MockSigningAgentClient) ClientInit(register *api.QredoRegisterInitRequest, ref, apikey, b64PrivateKey string) (*api.QredoRegisterInitResponse, error) {
	m.ClientInitCalled = true
	m.LastRegisterRequest = register
	m.LastRef = ref
	m.LastApiKey = apikey
	m.Last64PrivateKey = b64PrivateKey
	return m.NextRegisterInitResponse, m.NextClientInitError
}

func (m *MockSigningAgentClient) ClientRegister(name string) (*api.ClientRegisterResponse, error) {
	m.ClientRegisterCalled = true
	m.LastRegisteredName = name
	return m.NextClientRegisterResponse, m.NextRegisterError
}

func (m *MockSigningAgentClient) ClientRegisterFinish(req *api.ClientRegisterFinishRequest, ref string) (*api.ClientRegisterFinishResponse, error) {
	m.ClientRegisterFinishCalled = true
	m.LastRegisterFinishRequest = req
	return m.NextRegisterFinishResponse, m.NextRegisterFinishError
}

func (m *MockSigningAgentClient) GetAgentID() string {
	m.GetAgentIDCalled = true

	return m.NextAgentID
}

func (m *MockSigningAgentClient) GetAgentName(agentID string) string {
	m.GetAgentNameCalled = true
	m.LastAgentID = agentID

	return m.NextAgentName
}

func (m *MockSigningAgentClient) ActionApprove(actionID string) error {
	m.ActionApproveCalled = true
	m.LastActionId = actionID
	m.Counter++
	return m.NextError
}

func (m *MockSigningAgentClient) ActionReject(actionID string) error {
	m.ActionRejectCalled = true
	m.LastRejectActionId = actionID
	return m.NextError
}

func (m *MockSigningAgentClient) SetSystemAgentID(agetID string) error {
	return nil
}

func (m *MockSigningAgentClient) GetSystemAgentID() string {
	m.GetSystemAgentIDCalled = true
	return m.NextAgentID
}

func (m *MockSigningAgentClient) GetAgentZKPOnePass() ([]byte, error) {
	m.Counter++
	m.GetAgentZKPOnePassCalled = true
	return m.NextZKPOnePass, m.NextError
}

func (m *MockSigningAgentClient) ReadAction(string, ServeCB) *feed {
	return nil
}
