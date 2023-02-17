package autoapprover

import (
	"errors"
	"testing"

	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"

	"github.com/stretchr/testify/assert"
)

func TestActionManage_Approve_shouldnt_handle_action(t *testing.T) {
	//Arrange
	syncronizerMock := &mockActionSyncronizer{
		NextShouldHandle: false,
	}
	coreMock := &lib.MockSigningAgentClient{}
	sut := NewActionManager(coreMock, syncronizerMock, util.NewTestLogger(), true)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.Nil(t, res)
	assert.True(t, syncronizerMock.ShouldHandleActionCalled)
	assert.Equal(t, "some test action id", syncronizerMock.LastActionId)
	assert.False(t, coreMock.ActionApproveCalled)
}

func TestActionManage_Approve_fails_to_acquire_lock(t *testing.T) {
	//Arrange
	syncronizerMock := &mockActionSyncronizer{
		NextShouldHandle: true,
		NextLockError:    errors.New("some lock error"),
	}
	coreMock := &lib.MockSigningAgentClient{}
	sut := NewActionManager(coreMock, syncronizerMock, util.NewTestLogger(), true)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.NotNil(t, res)
	assert.Equal(t, "some lock error", res.Error())
	assert.True(t, syncronizerMock.AcquireLockCalled)
	assert.False(t, coreMock.ActionApproveCalled)
}

func TestActionManage_Approve_approves(t *testing.T) {
	//Arrange
	syncronizerMock := &mockActionSyncronizer{
		NextShouldHandle: true,
		NextReleaseError: errors.New("some unlock error"),
	}
	coreMock := &lib.MockSigningAgentClient{}
	sut := NewActionManager(coreMock, syncronizerMock, util.NewTestLogger(), true)

	//Act
	res := sut.Approve("some test action id")

	//Assert
	assert.Nil(t, res)
	assert.True(t, syncronizerMock.ReleaseCalled)
	assert.True(t, coreMock.ActionApproveCalled)
}

func TestActionManage_Reject_returns_error(t *testing.T) {
	//Arrange
	coreMock := &lib.MockSigningAgentClient{
		NextError: errors.New("some reject error"),
	}
	sut := NewActionManager(coreMock, nil, util.NewTestLogger(), true)

	//Act
	err := sut.Reject("some test action id")

	//Assert
	assert.NotNil(t, err)
	assert.Equal(t, "some reject error", err.Error())
	assert.True(t, coreMock.ActionRejectCalled)
	assert.Equal(t, "some test action id", coreMock.LastRejectActionId)
}
