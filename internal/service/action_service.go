package service

import (
	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/hub/message"
	"go.uber.org/zap"
)

type ActionService interface {
	Reject(actionID string) error
	Approve(actionID string) error
}

func NewActionService(syncronizer action.ActionSync, log *zap.SugaredLogger, loadBalancingEnabled bool, messageCache message.CacheRemover, signer action.Signer) ActionService {
	return &actionSrv{
		syncronizer:          syncronizer,
		log:                  log,
		loadBalancingEnabled: loadBalancingEnabled,
		messageCache:         messageCache,
		signer:               signer,
	}
}

type actionSrv struct {
	syncronizer          action.ActionSync
	log                  *zap.SugaredLogger
	loadBalancingEnabled bool
	messageCache         message.CacheRemover
	signer               action.Signer
}

// Approve the action for the given actionID
func (a *actionSrv) Approve(actionID string) error {
	a.log.Infof("Action Service: approving action `%s`", actionID)
	return a.act(actionID, true)
}

// Reject the action for the given actionID
func (a *actionSrv) Reject(actionID string) error {
	a.log.Infof("Action Service: rejecting action `%s`", actionID)
	return a.act(actionID, false)
}

func (a *actionSrv) act(actionID string, approve bool) error {
	if a.loadBalancingEnabled {
		if !a.syncronizer.ShouldHandleAction(actionID) {
			a.log.Infof("Action Service: action `%s` was already handled!", actionID)
			return nil
		}

		if err := a.syncronizer.AcquireLock(); err != nil {
			a.log.Errorf("Action Service: lock acquire err: %v, actionID `%s`", err, actionID)
			return err
		}
		defer func() {
			if err := a.syncronizer.Release(actionID); err != nil {
				a.log.Errorf("Action Service: lock release err: %v, actionID `%s`", err, actionID)
			}
		}()
	}

	var err error
	if approve {
		err = a.signer.ActionApprove(actionID)
	} else {
		err = a.signer.ActionReject(actionID)
	}

	if err != nil {
		return err
	}

	if a.messageCache != nil {
		a.messageCache.RemoveMessage(actionID)
	}

	return nil
}
