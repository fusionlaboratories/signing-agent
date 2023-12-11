package action

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/hub/message"
)

var ctxBackground = context.Background()

// ActionSync provides functionality to manage the approval of an action when load balancing is enabled
type ActionSync interface {
	ShouldHandleAction(actionID string) bool
	AcquireLock() error
	Release(actionID string) error
}

type syncI interface {
	NewMutex(name string, options ...redsync.Option) *redsync.Mutex
}

type mutex interface {
	Lock() error
	Unlock() (bool, error)
}

type syncronize struct {
	cache            message.KVStore
	mutex            mutex
	sync             syncI
	cfgLoadBalancing *config.LoadBalancing
}

// NewSyncronizer returns a new ActionSyncronizer that's an instance of syncronize
func NewSyncronizer(conf *config.LoadBalancing, cache message.KVStore, sync syncI) ActionSync {
	return &syncronize{
		cfgLoadBalancing: conf,
		cache:            cache,
		sync:             sync,
	}
}

// ShouldHandleAction returns true if the action wasn't already picked up by another agent
func (a *syncronize) ShouldHandleAction(actionID string) bool {
	if err := a.cache.Get(ctxBackground, a.getKey(actionID)).Err(); err == nil {
		return false
	}

	// set the mutex to lock the action
	a.mutex = a.sync.NewMutex(actionID)
	return true
}

// AcquireLock locks the mutex set for the action to be handled
func (a *syncronize) AcquireLock() error {
	if err := a.mutex.Lock(); err != nil {
		time.Sleep(time.Duration(a.cfgLoadBalancing.OnLockErrorTimeOutMs) * time.Millisecond)
		return err
	}

	return nil
}

// Release unlocks the mutex and sets the action id in the cache to signal it was already handled
func (a *syncronize) Release(actionID string) error {
	_, err := a.mutex.Unlock()
	a.cache.Set(ctxBackground, a.getKey(actionID), 1, time.Duration(a.cfgLoadBalancing.ActionIDExpirationSec)*time.Second)

	return err
}

func (a *syncronize) getKey(actionID string) string {
	return "action_v2:" + actionID
}
