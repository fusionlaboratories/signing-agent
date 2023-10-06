package autoapprover

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/hub"
	"github.com/qredo/signing-agent/internal/util"
	"github.com/test-go/testify/assert"
	"go.uber.org/goleak"
)

var ignoreOpenCensus = goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")

func TestAutoApprover_Listen_fails_to_unmarshal(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	signerMock := &action.MockSigner{}
	sut := &autoActionApprover{
		signer:        signerMock,
		HubFeedClient: hub.NewHubFeedClient(true),
		log:           util.NewTestLogger(),
	}
	defer sut.Stop()
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
	assert.False(t, signerMock.ApproveActionMessageCalled)
}

func TestAutoApprover_handleMessage_action_expired(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	signerMock := &action.MockSigner{}
	sut := &autoActionApprover{
		log:    util.NewTestLogger(),
		signer: signerMock,
	}
	bytes, _ := json.Marshal(defs.ActionInfo{
		ExpireTime: 12360,
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.Nil(t, sut.lastError)
	assert.False(t, signerMock.ApproveActionMessageCalled)
}

func TestAutoApprover_handleMessage_action_notInPending(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	signerMock := &action.MockSigner{}
	sut := &autoActionApprover{
		log:    util.NewTestLogger(),
		signer: signerMock,
	}
	bytes, _ := json.Marshal(defs.ActionInfo{
		ExpireTime: time.Now().Add(time.Hour).Unix(),
		Status:     3,
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.Nil(t, sut.lastError)
	assert.False(t, signerMock.ApproveActionMessageCalled)
}

func TestAutoApprover_handleMessage_shouldnt_handle_action(t *testing.T) {
	//Arrange
	syncronizerMock := &action.MockActionSyncronizer{}
	signerMock := &action.MockSigner{}
	sut := autoActionApprover{
		log:                  util.NewTestLogger(),
		loadBalancingEnabled: true,
		syncronizer:          syncronizerMock,
		signer:               signerMock}
	bytes, _ := json.Marshal(defs.ActionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
		Status:     defs.StatusPending,
	})

	//Act
	sut.handleMessage(bytes)

	//Assert
	assert.True(t, syncronizerMock.ShouldHandleActionCalled)
	assert.Equal(t, "actionid", syncronizerMock.LastActionId)
	assert.False(t, signerMock.ApproveActionMessageCalled)
}

func TestAutoApprover_handleMessage_fails_to_lock(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	syncronizerMock := &action.MockActionSyncronizer{
		NextLockError:    errors.New("some lock error"),
		NextShouldHandle: true,
	}
	sut := autoActionApprover{
		log:                  util.NewTestLogger(),
		loadBalancingEnabled: true,
		syncronizer:          syncronizerMock,
	}

	bytes, _ := json.Marshal(defs.ActionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
		Status:     defs.StatusPending,
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
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	syncronizerMock := &action.MockActionSyncronizer{
		NextReleaseError: errors.New("some release error"),
	}
	signerMock := &action.MockSigner{}
	sut := autoActionApprover{
		log:                  util.NewTestLogger(),
		loadBalancingEnabled: true,
		syncronizer:          syncronizerMock,
		signer:               signerMock,
	}
	action := defs.ActionInfo{
		ID:         "actionid",
		ExpireTime: time.Now().Add(time.Minute).Unix(),
		Status:     defs.StatusPending,
		Messages:   [][]byte{[]byte("some message")},
	}

	//Act
	sut.handleAction(action)

	//Assert
	assert.True(t, syncronizerMock.AcquireLockCalled)
	assert.True(t, syncronizerMock.ReleaseCalled)
	assert.True(t, signerMock.ApproveActionMessageCalled)
	assert.Equal(t, "actionid", signerMock.LastActionId)
	assert.Equal(t, []byte("some message"), signerMock.LastMessage)
}

func TestAutoApprover_approveAction_retries_to_approve(t *testing.T) {
	//Arrange
	signerMock := &action.MockSigner{
		NextError: errors.New("some error"),
	}
	sut := &autoActionApprover{
		signer: signerMock,
		cfgAutoApproval: config.AutoApprove{
			RetryIntervalMax: 3,
			RetryInterval:    1,
		},
		log: util.NewTestLogger(),
	}

	//Act
	sut.approveAction("some action id", []byte("some message"))

	//Assert
	assert.True(t, signerMock.ApproveActionMessageCalled)
	assert.Equal(t, "some action id", signerMock.LastActionId)
	assert.True(t, signerMock.Counter > 1)
}
