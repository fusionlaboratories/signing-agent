package actions

import (
	"github.com/qredo/signing-agent/hub/message"
	"github.com/qredo/signing-agent/lib"

	"go.uber.org/zap"
)

// Action manager provides functionality to approve and reject an action given its id
type ActionManager interface {
	Approve(actionID string) error
	Reject(actionID string) error
}

type actionManage struct {
	core                 lib.SigningAgentClient
	syncronizer          ActionSyncronizer
	log                  *zap.SugaredLogger
	loadBalancingEnabled bool
	messageCache         message.CacheRemover
}

// NewActionManager return an ActionManager that's an instance of actionManage
func NewActionManager(core lib.SigningAgentClient, syncronizer ActionSyncronizer, log *zap.SugaredLogger, loadBalancingEnabled bool, messageCache message.CacheRemover) ActionManager {
	return &actionManage{
		core:                 core,
		syncronizer:          syncronizer,
		log:                  log,
		loadBalancingEnabled: loadBalancingEnabled,
		messageCache:         messageCache,
	}
}

// Approve the action for the given actionID
func (a *actionManage) Approve(actionID string) error {
	a.log.Debugf("manually approving action `%s`", actionID)
	return a.act(actionID, a.core.ActionApprove)
}

// Reject the action for the given actionID
func (a *actionManage) Reject(actionID string) error {
	a.log.Debugf("manually rejecting action `%s`", actionID)
	return a.act(actionID, a.core.ActionReject)
}

func (a *actionManage) act(actionID string, actFunc func(string) error) error {
	if a.loadBalancingEnabled {
		if !a.syncronizer.ShouldHandleAction(actionID) {
			a.log.Debugf("action [%s] was already handled!", actionID)
			return nil
		}

		if err := a.syncronizer.AcquireLock(); err != nil {
			a.log.Errorf("%v action-id %s", err, actionID)
			return err
		}
		defer func() {
			if err := a.syncronizer.Release(actionID); err != nil {
				a.log.Errorf("%v action-id %s", err, actionID)
			}
		}()
	}

	if err := actFunc(actionID); err != nil {
		return err
	}

	if a.messageCache != nil {
		a.messageCache.RemoveMessage(actionID)
	}

	return nil
}
