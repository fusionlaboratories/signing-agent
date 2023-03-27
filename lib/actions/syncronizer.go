package actions

import (
	"context"
	"time"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/lib/store"
)

const keyPrefix = "action:"

var rCtx = context.Background()

// ActionSyncronizer provides functionality to manage the approval of an action when load balancing is enabled
type ActionSyncronizer interface {
	ShouldHandleAction(actionID string) bool
	AcquireLock() error
	Release(actionID string) error
}

type syncronize struct {
	cache            store.KVStore
	mutex            mutex
	sync             syncI
	cfgLoadBalancing *config.LoadBalancing
}

// NewSyncronizer returns a new ActionSyncronizer that's an instance of syncronize
func NewSyncronizer(conf *config.LoadBalancing, cache store.KVStore, sync syncI) ActionSyncronizer {
	return &syncronize{
		cfgLoadBalancing: conf,
		cache:            cache,
		sync:             sync,
	}
}

// ShouldHandleAction returns true if the action wasn't already picked up by another agent
func (a *syncronize) ShouldHandleAction(actionID string) bool {
	if err := a.cache.Get(rCtx, a.getKey(actionID)).Err(); err == nil {
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
	a.cache.Set(rCtx, a.getKey(actionID), 1, time.Duration(a.cfgLoadBalancing.ActionIDExpirationSec)*time.Second)

	return err
}

func (a *syncronize) getKey(actionID string) string {
	return keyPrefix + actionID
}
