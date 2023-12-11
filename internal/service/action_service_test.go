package service

import (
	"errors"
	"testing"

	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/hub/message"
	"github.com/test-go/testify/assert"
)

func TestActionService_Approve_shouldnt_handle_action(t *testing.T) {
	//Arrange
	syncronizerMock := &action.MockActionSyncronizer{
		NextShouldHandle: false,
	}
	signerMock := &action.MockSigner{}

	sut := NewActionService(syncronizerMock, testLog, true, nil, signerMock)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.Nil(t, res)
	assert.True(t, syncronizerMock.ShouldHandleActionCalled)
	assert.Equal(t, "some test action id", syncronizerMock.LastActionId)
	assert.False(t, signerMock.ActionApproveCalled)
}

func TestActionService_Approve_fails_to_acquire_lock(t *testing.T) {
	//Arrange
	syncronizerMock := &action.MockActionSyncronizer{
		NextShouldHandle: true,
		NextLockError:    errors.New("some lock error"),
	}
	signerMock := &action.MockSigner{}
	sut := NewActionService(syncronizerMock, testLog, true, nil, signerMock)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.NotNil(t, res)
	assert.Equal(t, "some lock error", res.Error())
	assert.True(t, syncronizerMock.AcquireLockCalled)
	assert.False(t, signerMock.ActionApproveCalled)
}

func TestActionService_Approve_approves(t *testing.T) {
	//Arrange
	syncronizerMock := &action.MockActionSyncronizer{
		NextShouldHandle: true,
		NextReleaseError: errors.New("some unlock error"),
	}
	signerMock := &action.MockSigner{}
	sut := NewActionService(syncronizerMock, testLog, true, nil, signerMock)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.Nil(t, res)
	assert.True(t, syncronizerMock.ReleaseCalled)
	assert.True(t, signerMock.ActionApproveCalled)
	assert.Equal(t, "some test action id", signerMock.LastActionId)
}

func TestActionService_Reject_returns_error_doesnt_remove_from_cache(t *testing.T) {
	//Arrange
	signerMock := &action.MockSigner{
		NextError: errors.New("some reject error"),
	}
	cacheMock := &message.MockCache{}
	sut := NewActionService(nil, testLog, false, cacheMock, signerMock)

	//Act
	err := sut.Reject("some test action id")

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "some reject error", err.Error())
	assert.True(t, signerMock.ActionRejectCalled)
	assert.Equal(t, "some test action id", signerMock.LastActionId)
	assert.False(t, cacheMock.RemoveMessageCalled)
}

func TestActionService_Reject_rejects_removes_from_cache(t *testing.T) {
	//Arrange
	signerMock := &action.MockSigner{}
	cacheMock := &message.MockCache{}
	sut := NewActionService(nil, testLog, false, cacheMock, signerMock)

	//Act
	err := sut.Reject("some test action id")

	//Assert
	assert.Nil(t, err)
	assert.True(t, signerMock.ActionRejectCalled)
	assert.Equal(t, "some test action id", signerMock.LastActionId)
	assert.True(t, cacheMock.RemoveMessageCalled)
	assert.Equal(t, "some test action id", cacheMock.LastID)
}
