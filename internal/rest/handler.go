package rest

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/defs"
)

func (a Router) RegisterAgent(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (any, error) {
	data := &api.AgentRegisterRequest{}
	if err := a.decode(data, r); err != nil {
		a.log.Debugf("failed to decode register request, %v", err)
		return nil, err
	}

	resp, err := a.agentService.RegisterAgent(data)
	if err != nil {
		return nil, err
	}

	a.log.Info("agent registered, starting the service")

	if err := a.agentService.Start(); err != nil {
		a.log.Errorf("failed to start the agent service, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to start the agent service. Please restart")
	}

	return resp, nil
}

func (a Router) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (any, error) {
	a.agentService.RegisterClientFeed(w, r)
	return nil, nil
}

func (a Router) GetClient(_ *defs.RequestContext, w http.ResponseWriter, _ *http.Request) (any, error) {
	return a.agentService.GetAgentDetails()
}

func (a Router) ActionApprove(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (any, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}

	if err := a.actionService.Approve(actionID); err != nil {
		return nil, err
	}

	return api.ActionResponse{
		ActionID: actionID,
		Status:   "approved",
	}, nil
}

func (a Router) ActionReject(_ *defs.RequestContext, _ http.ResponseWriter, r *http.Request) (any, error) {
	actionID := mux.Vars(r)["action_id"]
	actionID = strings.TrimSpace(actionID)
	if actionID == "" {
		return nil, defs.ErrBadRequest().WithDetail("empty actionID")
	}

	if err := a.actionService.Reject(actionID); err != nil {
		return nil, err
	}

	return api.ActionResponse{
		ActionID: actionID,
		Status:   "rejected",
	}, nil
}

func (a Router) HealthCheckVersion(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (any, error) {
	return a.version, nil
}

func (a Router) HealthCheckConfig(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (any, error) {
	return a.config, nil
}

func (a Router) HealthCheckStatus(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (any, error) {
	return a.agentService.GetWebsocketStatus(), nil
}
