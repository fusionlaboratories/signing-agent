package autoapprover

import (
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
}

// NewActionManager return an ActionManager that's an instance of actionManage
func NewActionManager(core lib.SigningAgentClient, syncronizer ActionSyncronizer, log *zap.SugaredLogger, loadBalancingEnabled bool) ActionManager {
	return &actionManage{
		core:                 core,
		syncronizer:          syncronizer,
		log:                  log,
		loadBalancingEnabled: loadBalancingEnabled,
	}
}

// Approve the action for the given actionID
func (a *actionManage) Approve(actionID string) error {
	if a.loadBalancingEnabled {
		if !a.syncronizer.ShouldHandleAction(actionID) {
			a.log.Debugf("action [%v] was already approved!", actionID)
			return nil
		}

		if err := a.syncronizer.AcquireLock(); err != nil {
			a.log.Errorf("%v action-id %v", err, actionID)
			return err
		}
		defer func() {
			if err := a.syncronizer.Release(actionID); err != nil {
				a.log.Errorf("%v action-id %v", err, actionID)
			}
		}()
	}

	return a.core.ActionApprove(actionID)
}

// Reject the action for the given actionID
func (a *actionManage) Reject(actionID string) error {
	return a.core.ActionReject(actionID)
}
