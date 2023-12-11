package auth

import (
	"net/http"
)

type MockHeaderProvider struct {
	InitiateCalled      bool
	GetAuthHeaderCalled bool
	StopCalled          bool

	LastWorkspaceID  string
	LastApiKeySecret string
	LastApiKeyID     string

	NextError  error
	NextHeader http.Header

	Counter int
}

func (m *MockHeaderProvider) Initiate(workspaceID, apiKeySecret, apiKeyID string) error {
	m.InitiateCalled = true
	m.LastApiKeyID = apiKeyID
	m.LastApiKeySecret = apiKeySecret
	m.LastWorkspaceID = workspaceID

	return m.NextError
}

func (m *MockHeaderProvider) GetAuthHeader() http.Header {
	m.GetAuthHeaderCalled = true
	m.Counter++
	return m.NextHeader
}

func (m *MockHeaderProvider) Stop() {
	m.StopCalled = true
}
