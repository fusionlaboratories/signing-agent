package clients

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qredo/signing-agent/clientfeed"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
	"github.com/test-go/testify/assert"
	"go.uber.org/goleak"
	"go.uber.org/zap"
)

var testLog = util.NewTestLogger()

func TestClientsManager_Start_agent_not_registered_doesnt_run_hub(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	mockCore := lib.NewMockSigningAgentClient("")
	sut := NewManager(mockCore, mockFeedHub, testLog,
		nil, nil, nil)

	//Act
	sut.Start()

	//Assert
	assert.False(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestClientsManager_Start_autoapprover_not_enabled_hub_doesnt_run(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: false,
	}
	mockCore := lib.NewMockSigningAgentClient("agent id")
	sut := NewManager(mockCore, mockFeedHub, testLog,
		&config.Config{}, nil, nil)

	//Act
	sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestClientsManager_Start_autoapprover_enabled_hub_doesnt_run(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: false,
	}
	mockCore := lib.NewMockSigningAgentClient("agent id")
	sut := NewManager(mockCore, mockFeedHub, testLog,
		&config.Config{
			AutoApprove: config.AutoApprove{
				Enabled: true,
			},
		}, nil, nil)

	//Act
	sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestClientsManager_Start_run_hub_autoapprove_not_enabled(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("agent id")
	sut := NewManager(mockCore, mockFeedHub, testLog,
		&config.Config{}, nil, nil)

	//Act
	sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestClientsManager_Start_registers_auto_approval(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}
	mockCore := lib.NewMockSigningAgentClient("valid_agentID")

	sut := NewManager(mockCore, mockFeedHub, testLog, &config.Config{
		AutoApprove: config.AutoApprove{
			Enabled: true,
		},
	}, nil, nil)

	//Act
	sut.Start()

	//Assert
	assert.True(t, mockFeedHub.RunCalled)
	assert.True(t, mockCore.GetSystemAgentIDCalled)
	assert.True(t, mockFeedHub.RegisterClientCalled)

	autoApprover := mockFeedHub.LastRegisteredClient
	assert.NotNil(t, autoApprover)
	assert.True(t, autoApprover.IsInternal)
	close(autoApprover.Feed)
}

func TestClientsManager_Stop_stops_feedhub(t *testing.T) {
	//Arrange
	mockFeedHub := &mockFeedHub{}
	sut := NewManager(nil, mockFeedHub, testLog, &config.Config{}, nil, nil)

	//Act
	sut.Stop()

	//Assert
	assert.True(t, mockFeedHub.StopCalled)
}

func TestClientsManager_RegisterClientFeed_hub_not_running(t *testing.T) {
	// Arrange
	mockFeedHub := &mockFeedHub{}
	sut := NewManager(nil, mockFeedHub, testLog, &config.Config{}, nil, nil)

	// Act
	sut.RegisterClientFeed(nil, nil)

	// Assert
	assert.False(t, mockFeedHub.RegisterClientCalled)
}

func TestClientsManager_RegisterClientFeed_upgrade_fails(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}

	mockUpgrader := &hub.MockWebsocketUpgrader{
		NextError: errors.New("some upgrade error"),
	}
	sut := NewManager(nil, mockFeedHub, testLog, &config.Config{}, mockUpgrader, nil)

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

func TestClientsManager_RegisterClientFeed_client_registered(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockFeedHub := &mockFeedHub{
		NextRun: true,
	}

	mockUpgrader := &hub.MockWebsocketUpgrader{
		NextWebsocketConnection: &hub.MockWebsocketConnection{},
	}

	mockClientFeed := &mockClientFeed{
		NextFeedClient: &hub.FeedClient{},
	}

	sut := clientsManager{
		feedHub:  mockFeedHub,
		log:      testLog,
		upgrader: mockUpgrader,
		config:   &config.Config{},
		newClientFeedFunc: func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister clientfeed.UnregisterFunc, config *config.WebSocketConfig) clientfeed.ClientFeed {
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
