package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/rest/version"
)

type mockSourceStats struct {
	GetFeedUrlCalled    bool
	GetReadyStateCalled bool
	NextFeedUrl         string
	NextReadyState      string
}

func (m *mockSourceStats) GetFeedUrl() string {
	m.GetFeedUrlCalled = true
	return m.NextFeedUrl
}

func (m *mockSourceStats) GetReadyState() string {
	m.GetReadyStateCalled = true
	return m.NextReadyState
}

type mockConnectedClients struct {
	GetExternalFeedClientsCalled bool
	NextConnectedClients         int
}

func (m *mockConnectedClients) GetExternalFeedClients() int {
	m.GetExternalFeedClientsCalled = true
	return m.NextConnectedClients
}

func TestHealthCheckHandler_HealthCheckStatus(t *testing.T) {
	//Arrange
	sourceMock := &mockSourceStats{
		NextFeedUrl:    "some feel url",
		NextReadyState: "some ready status",
	}

	connClientsMock := &mockConnectedClients{
		NextConnectedClients: 7,
	}

	handler := NewHealthCheckHandler(sourceMock, nil, nil, connClientsMock, "some local feed")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckStatus(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)

	data, _ := json.Marshal(response)
	assert.NotEmpty(t, string(data))

	assert.True(t, sourceMock.GetFeedUrlCalled)
	assert.True(t, sourceMock.GetReadyStateCalled)
	assert.True(t, connClientsMock.GetExternalFeedClientsCalled)
}

func TestHealthCheckHandler_HealthCheckVersion(t *testing.T) {
	//Arrange
	version := &version.Version{
		BuildVersion: "some build version",
		BuildType:    "some build type",
		BuildDate:    "some build date",
	}
	handler := NewHealthCheckHandler(nil, version, nil, nil, "")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckVersion(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)

	data, _ := json.Marshal(response)
	assert.Equal(t, "{\"buildVersion\":\"some build version\",\"buildType\":\"some build type\",\"buildDate\":\"some build date\"}", string(data))
}

func TestHealthCheckHandler_HealthCheckConfig(t *testing.T) {
	//Arrange
	config := &config.Config{
		Base: config.Base{
			PIN:      25,
			QredoAPI: "some url",
		},
		HTTP: config.HttpSettings{
			Addr: "some address",
		},
	}
	handler := NewHealthCheckHandler(nil, nil, config, nil, "")

	req, _ := http.NewRequest("GET", "/path", nil)
	rr := httptest.NewRecorder()

	//Act
	response, err := handler.HealthCheckConfig(nil, rr, req)

	//Assert
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, rr.Code)

	data, _ := json.Marshal(response)
	assert.NotEmpty(t, string(data))
}
