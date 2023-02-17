package lib

import (
	"encoding/hex"
	"net/http"

	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/util"
)

func (h *signingAgent) ActionApprove(actionID string) error {
	agentID := h.store.GetSystemAgentID()
	if agentID == "" {
		return defs.ErrNotFound().WithDetail("agentID")
	}
	agent := h.store.GetAgent(agentID)
	if agent == nil {
		return defs.ErrNotFound().WithDetail("agent")
	}

	zkpOnePass, err := util.ZKPOnePass(agent.ZKPID, agent.ZKPToken, h.cfg.Base.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))
	messagesResp := &api.CoreClientServiceActionMessagesResponse{}
	if err = h.htc.Request(http.MethodGet, util.URLActionMessages(h.cfg.Base.QredoAPI, actionID), nil, messagesResp, header); err != nil {
		return err
	}

	if messagesResp.Messages == nil || len(messagesResp.Messages) == 0 {
		return defs.ErrNotFound().WithDetail("messages")
	}

	signatures := make([]string, len(messagesResp.Messages))

	for i, m := range messagesResp.Messages {
		msg, err := hex.DecodeString(m)
		if err != nil || len(msg) == 0 {
			return err
		}

		signature, err := util.BLSSign(agent.BLSSeed, msg)
		if err != nil {
			return err
		}
		signatures[i] = hex.EncodeToString(signature)
	}

	zkpOnePass, err = util.ZKPOnePass(agent.ZKPID, agent.ZKPToken, h.cfg.Base.PIN)
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	req := &api.CoreClientServiceActionApproveRequest{
		Signatures: signatures,
	}
	header = http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))
	if err = h.htc.Request(http.MethodPut, util.URLActionApprove(h.cfg.Base.QredoAPI, actionID), req, nil, header); err != nil {
		return err
	}

	return nil
}

func (h *signingAgent) ActionReject(actionID string) error {
	zkpOnePass, err := h.GetAgentZKPOnePass()
	if err != nil {
		return errors.Wrap(err, "get zkp token")
	}

	header := http.Header{}
	header.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))

	if err = h.htc.Request(http.MethodDelete, util.URLActionReject(h.cfg.Base.QredoAPI, actionID), nil, nil, header); err != nil {
		return err
	}

	return nil
}
