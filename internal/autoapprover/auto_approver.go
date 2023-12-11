// Package autoapprover provides a mechanism to receive action information as bytes.
// The action data is analyzed and if it meets the requirements, the action is approved.
// It supports approval retrying based on defined intervals

package autoapprover

import (
	"encoding/json"
	"sync"

	"go.uber.org/zap"

	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/hub"
)

type AutoApprover interface {
	Listen(wg *sync.WaitGroup)
	GetFeedClient() *hub.HubFeedClient
	Stop()
}

type autoActionApprover struct {
	hub.HubFeedClient
	log             *zap.SugaredLogger
	cfgAutoApproval config.AutoApprove

	syncronizer          action.ActionSync
	lastError            error
	loadBalancingEnabled bool
	signer               action.Signer
}

// NewAutoApprover returns a new *AutoApprover instance initialized with the provided parameters
// The AutoApprover has an internal FeedClient which means it will be stopped when the service stops
// or the Feed channel is closed on the sender side
func NewAutoApprover(log *zap.SugaredLogger, config config.Config, syncronizer action.ActionSync, signer action.Signer) AutoApprover {
	return &autoActionApprover{
		HubFeedClient:        hub.NewHubFeedClient(true),
		log:                  log,
		cfgAutoApproval:      config.AutoApprove,
		syncronizer:          syncronizer,
		loadBalancingEnabled: config.LoadBalancing.Enable,
		signer:               signer,
	}
}

// Listen is constantly listening for messages on the Feed channel.
// The Feed channel is always closed by the sender. When this happens, the AutoApprover stops
func (a *autoActionApprover) Listen(wg *sync.WaitGroup) {
	a.log.Debug("AutoApprover: listening")
	wg.Done()

	for {
		if message, ok := <-a.Feed; !ok {
			//channel was closed by the sender
			a.log.Info("AutoApprover: stopped")
			return
		} else {
			go a.handleMessage(message)
		}
	}
}

func (a *autoActionApprover) Stop() {
	a.log.Debug("AutoApprover: stopping")
	close(a.Feed)
}

func (a *autoActionApprover) GetFeedClient() *hub.HubFeedClient {
	return &a.HubFeedClient
}

func (a *autoActionApprover) handleMessage(message []byte) {
	action := defs.ActionInfo{}
	if err := json.Unmarshal(message, &action); err == nil {
		if !action.IsExpired() {
			if action.Status == defs.StatusPending {
				if a.shouldHandleAction(action.ID) {
					a.handleAction(action)
				}
			} else {
				a.log.Infof("AutoApprover: action `%s` status not pending", action.ID)
			}
		} else {
			a.log.Infof("AutoApprover: action `%s` has expired", action.ID)
		}
	} else {
		a.log.Errorf("AutoApprover: fail to unmarshal the message `%v`, err: %v", string(message), err)
		a.lastError = err
	}
}

func (a *autoActionApprover) shouldHandleAction(actionId string) bool {
	if a.loadBalancingEnabled {
		//check if the action was already picked up by another signing agent
		if !a.syncronizer.ShouldHandleAction(actionId) {
			a.log.Debugf("AutoApprover: action `%s` was already approved!", actionId)
			return false
		}
	}

	return true
}

func (a *autoActionApprover) handleAction(action defs.ActionInfo) {
	if a.loadBalancingEnabled {
		if err := a.syncronizer.AcquireLock(); err != nil {
			a.log.Debugf("AutoApprover, mutex lock err: %v, action `%s`", err, action.ID)
			return
		}
		defer func() {
			if err := a.syncronizer.Release(action.ID); err != nil {
				a.log.Debugf("AutoApprover, mutex unlock err: %v, action `%s`", err, action.ID)
			}
		}()
	}

	a.approveAction(action.ID, action.Messages[0])
}

func (a *autoActionApprover) approveAction(actionId string, message []byte) {
	timer := newRetryTimer(a.cfgAutoApproval.RetryInterval, a.cfgAutoApproval.RetryIntervalMax)
	for {
		if err := a.signer.ApproveActionMessage(actionId, message); err == nil {
			a.log.Infof("AutoApprover: action `%s` approved automatically", actionId)
			return
		} else {
			a.log.Errorf("AutoApprover: approval failed for action `%s`, err: %v", actionId, err)

			if timer.isTimeOut() {
				a.log.Warnf("AutoApprover: auto action approve timed out for action `%s`", actionId)
				return
			}

			a.log.Warnf("AutoApprover: auto approve action is repeated for action `%s` ", actionId)
			timer.retry()
		}
	}
}
