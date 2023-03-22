package autoapprover

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"

	"github.com/test-go/testify/assert"
	"go.uber.org/goleak"
)

type mockActionSyncronizer struct {
	ShouldHandleActionCalled bool
	AcquireLockCalled        bool
	ReleaseCalled            bool
	LastActionId             string
	NextShouldHandle         bool
	NextLockError            error
	NextReleaseError         error
}

func (m *mockActionSyncronizer) ShouldHandleAction(actionID string) bool {
	m.ShouldHandleActionCalled = true
	m.LastActionId = actionID
	return m.NextShouldHandle
}
func (m *mockActionSyncronizer) AcquireLock() error {
	m.AcquireLockCalled = true
	return m.NextLockError
}
func (m *mockActionSyncronizer) Release(actionID string) error {
	m.ReleaseCalled = true
	m.LastActionId = actionID
	return m.NextReleaseError
}

func TestAutoApprover_Listen_fails_to_unmarshal(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_core := &lib.MockSigningAgentClient{}
	sut := &AutoApprover{
		core:       mock_core,
		FeedClient: hub.NewFeedClient(true),
		log:        util.NewTestLogger(),
	}
	defer close(sut.Feed)
	var wg sync.WaitGroup
	wg.Add(1)
	go sut.Listen(&wg)
	wg.Wait()

	//Act
	sut.Feed <- []byte("")
	<-time.After(time.Second) //give it time to finish

	//Assert
	assert.NotNil(t, sut.lastError)
	assert.Equal(t, "unexpected end of JSON input", sut.lastError.Error())
	assert.False(t, mock_core.ActionApproveCalled)
}

func TestAutoApprover_handleMessage_action_expired(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	sut := &AutoApprover{
		log: util.NewTestLogger(),
	}
	bytes, _ := json.Marshal(actionInfo{
		ExpireTime: 12360,
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.Nil(t, sut.lastError)
}

func TestAutoApprover_handleMessage_shouldnt_handle_action(t *testing.T) {
	//Arrange
	syncronizerMock := &mockActionSyncronizer{}

	sut := NewAutoApprover(nil, util.NewTestLogger(), &config.Config{LoadBalancing: config.LoadBalancing{Enable: true}}, syncronizerMock)
	bytes, _ := json.Marshal(actionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.True(t, syncronizerMock.ShouldHandleActionCalled)
	assert.Equal(t, "actionid", syncronizerMock.LastActionId)
}

func TestAutoApprover_handleMessage_fails_to_lock(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	syncronizerMock := &mockActionSyncronizer{
		NextLockError:    errors.New("some lock error"),
		NextShouldHandle: true,
	}

	sut := NewAutoApprover(nil, util.NewTestLogger(), &config.Config{LoadBalancing: config.LoadBalancing{Enable: true}}, syncronizerMock)
	bytes, _ := json.Marshal(actionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
	})

	//Act
	sut.handleMessage(bytes)
	<-time.After(time.Second) //give it a second to process

	//Assert
	assert.True(t, syncronizerMock.AcquireLockCalled)
	assert.False(t, syncronizerMock.ReleaseCalled)
}

func TestAutoApprover_handleAction_acquires_lock_and_approves(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	syncronizerMock := &mockActionSyncronizer{
		NextReleaseError: errors.New("some release error"),
	}
	coreMock := &lib.MockSigningAgentClient{}
	sut := NewAutoApprover(coreMock, util.NewTestLogger(), &config.Config{LoadBalancing: config.LoadBalancing{Enable: true}}, syncronizerMock)
	action := actionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
	}

	//Act
	sut.handleAction(&action)

	//Assert
	assert.True(t, syncronizerMock.AcquireLockCalled)
	assert.True(t, syncronizerMock.ReleaseCalled)
	assert.True(t, coreMock.ActionApproveCalled)
	assert.Equal(t, "actionid", coreMock.LastActionId)
}

func TestAutoApprover_approveAction_retries_to_approve(t *testing.T) {
	//Arrange
	coreMock := &lib.MockSigningAgentClient{
		NextError: errors.New("some error"),
	}
	sut := &AutoApprover{
		core: coreMock,
		cfgAutoApproval: &config.AutoApprove{
			RetryIntervalMax: 3,
			RetryInterval:    1,
		},
		log: util.NewTestLogger(),
	}

	//Act
	sut.approveAction("some action id", "some agent id")

	//Assert
	assert.True(t, coreMock.ActionApproveCalled)
	assert.Equal(t, "some action id", coreMock.LastActionId)
	assert.True(t, coreMock.Counter > 1)
}
