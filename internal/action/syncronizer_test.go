package action

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/hub/message"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/stretchr/testify/assert"
)

type mutexMock struct {
	LockCalled   bool
	UnlockCalled bool
	NextError    error
	NextUnlock   bool
}

func (m *mutexMock) Lock() error {
	m.LockCalled = true
	return m.NextError
}
func (m *mutexMock) Unlock() (bool, error) {
	m.UnlockCalled = true
	return m.NextUnlock, m.NextError
}

type mockSync struct {
	NewMutexCalled bool
	LastName       string
	NextMutex      *redsync.Mutex
}

func (m *mockSync) NewMutex(name string, options ...redsync.Option) *redsync.Mutex {
	m.NewMutexCalled = true
	m.LastName = name
	return m.NextMutex
}

func TestSyncronize_ShouldHandleAction_already_handled(t *testing.T) {
	//Arrange
	cacheMock := &message.KVStoreMock{
		NextStringCmd: redis.NewStringCmd(context.Background()),
	}
	sut := NewSyncronizer(&config.LoadBalancing{Enable: true}, cacheMock, nil)

	//Act
	res := sut.ShouldHandleAction("test action Id")

	//Assert
	assert.False(t, res)
	assert.True(t, cacheMock.GetCalled)
	assert.Equal(t, "action_v2:test action Id", cacheMock.LastKey)
}

func TestSyncronize_ShouldHandleAction_sets_mutex(t *testing.T) {
	//Arrange
	stringCmd := redis.NewStringCmd(context.Background())
	stringCmd.SetErr(errors.New("some error"))
	cacheMock := &message.KVStoreMock{
		NextStringCmd: stringCmd,
	}
	syncMock := &mockSync{
		NextMutex: &redsync.Mutex{},
	}
	sut := &syncronize{
		cfgLoadBalancing: &config.LoadBalancing{Enable: true},
		cache:            cacheMock,
		sync:             syncMock}

	//Act
	res := sut.ShouldHandleAction("test action Id")

	//Assert
	assert.True(t, res)
	assert.True(t, cacheMock.GetCalled)
	assert.Equal(t, "action_v2:test action Id", cacheMock.LastKey)
	assert.True(t, syncMock.NewMutexCalled)
	assert.Equal(t, "test action Id", syncMock.LastName)
	assert.NotNil(t, sut.mutex)
}

func TestSyncronize_AcquireLock_fails_to_lock_returns_error(t *testing.T) {
	//Arrange
	mutexMock := &mutexMock{
		NextError: errors.New("some lock error"),
	}
	sut := &syncronize{
		cfgLoadBalancing: &config.LoadBalancing{OnLockErrorTimeOutMs: 2},
		mutex:            mutexMock}

	//Act
	res := sut.AcquireLock()

	//Assert
	assert.NotNil(t, res)
	assert.Equal(t, "some lock error", res.Error())
	assert.True(t, mutexMock.LockCalled)
}

func TestSyncronize_AcquireLock_locks(t *testing.T) {
	//Arrange
	mutexMock := &mutexMock{}
	sut := &syncronize{
		cfgLoadBalancing: &config.LoadBalancing{OnLockErrorTimeOutMs: 2},
		mutex:            mutexMock}

	//Act
	res := sut.AcquireLock()

	//Assert
	assert.Nil(t, res)
	assert.True(t, mutexMock.LockCalled)
}

func TestSyncronize_Release_returns_error(t *testing.T) {
	//Arrange
	mutexMock := &mutexMock{
		NextUnlock: false,
		NextError:  errors.New("some unlock error"),
	}
	mockCache := &message.KVStoreMock{
		NextStringCmd: redis.NewStringCmd(context.Background()),
	}
	sut := &syncronize{
		cfgLoadBalancing: &config.LoadBalancing{ActionIDExpirationSec: 2},
		mutex:            mutexMock,
		cache:            mockCache,
	}

	//Act
	res := sut.Release("test action id")

	//Assert
	assert.NotNil(t, res)
	assert.Equal(t, "some unlock error", res.Error())
	assert.True(t, mutexMock.UnlockCalled)
	assert.True(t, mockCache.SetCalled)
	assert.Equal(t, "action_v2:test action id", mockCache.LastKey)
	assert.Equal(t, 1, mockCache.LastValue)
	assert.Equal(t, 2*time.Second, mockCache.LastExpiration)
}
