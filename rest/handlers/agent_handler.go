package handlers

import (
	"net/http"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/lib/clients"
	"github.com/qredo/signing-agent/util"
)

type SigningAgentHandler struct {
	log  *zap.SugaredLogger
	core lib.SigningAgentClient

	localFeed    string
	decode       func(interface{}, *http.Request) error
	agentManager clients.AgentMng
}

// NewSigningAgentHandler instantiates and returns a new SigningAgentHandler object.
func NewSigningAgentHandler(agentManager clients.AgentMng, core lib.SigningAgentClient, log *zap.SugaredLogger, localFeed string) *SigningAgentHandler {
	return &SigningAgentHandler{
		agentManager: agentManager,
		log:          log,
		core:         core,
		localFeed:    localFeed,
		decode:       util.DecodeRequest,
	}
}

// RegisterAgent
//
// swagger:route POST /register client RegisterAgent
//
// # Register a new agent
//
// This will register the agent only if there is none already registered.
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// Responses:
//
// 200: AgentRegisterResponse
// 400: ErrorResponse description:Bad request
// 404: ErrorResponse description:Not found
// 500: ErrorResponse description:Internal error
func (h *SigningAgentHandler) RegisterAgent(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if h.core.GetSystemAgentID() != "" {
		return nil, defs.ErrBadRequest().WithDetail("AgentID already exist. You can not set new one.")
	}

	if response, err := h.register(r); err != nil {
		return nil, err
	} else {
		h.agentManager.Start()
		return response, nil
	}
}

// ClientFeed
//
// swagger:route GET /client/feed client ClientFeed
//
// # Get approval requests Feed (via websocket) from Qredo Backend
//
// This endpoint feeds approval requests coming from the Qredo Backend to the agent.
//
//	Produces:
//	- application/json
//
//	Schemes: ws, wss
//
// Responses:
// 200: ClientFeedResponse
func (h *SigningAgentHandler) ClientFeed(_ *defs.RequestContext, w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.agentManager.RegisterClientFeed(w, r)
	return nil, nil
}

// GetClient
//
// swagger:route GET /client client GetClient
//
// # Get information about the registered agent
//
// This endpoint retrieves the `agentID` and `feedURL` if an agent is registered.
//
// Produces:
//   - application/json
//
// Responses:
//
//	200: GetClientResponse
func (h *SigningAgentHandler) GetClient(_ *defs.RequestContext, w http.ResponseWriter, _ *http.Request) (interface{}, error) {
	agentID := h.core.GetAgentID()

	response := api.GetClientResponse{
		AgentName: h.core.GetAgentName(agentID),
		AgentID:   agentID,
		FeedURL:   h.localFeed,
	}

	return response, nil
}

func (h *SigningAgentHandler) register(r *http.Request) (interface{}, error) {
	registerRequest, err := h.validateRegisterRequest(r)
	if err != nil {
		return nil, err
	}

	registerResults, err := h.core.ClientRegister(registerRequest.Name) // Get BLS and EC public keys
	if err != nil {
		h.log.Debugf("error while trying to register the client [%s], err: %v", registerRequest.Name, err)
		return nil, err
	}

	initResults, err := h.initRegistration(registerResults, registerRequest)
	if err != nil {
		h.log.Debugf("error while trying to init the client registration, err: %v", err)
		return nil, err
	}

	if err := h.finishRegistration(initResults, registerResults.RefID); err != nil {
		return nil, err
	}

	response := api.AgentRegisterResponse{
		AgentID: initResults.AccountCode,
		FeedURL: h.localFeed,
	}

	return response, nil
}

func (h *SigningAgentHandler) validateRegisterRequest(r *http.Request) (*api.ClientRegisterRequest, error) {
	register := &api.ClientRegisterRequest{}
	if err := h.decode(register, r); err != nil {
		h.log.Debugf("failed to decode register request, %v", err)
		return nil, err
	}

	if err := register.Validate(); err != nil {
		h.log.Debugf("failed to validate register request, %v", err)
		return nil, defs.ErrBadRequest().WithDetail(err.Error())
	}

	return register, nil
}

func (h *SigningAgentHandler) initRegistration(register *api.ClientRegisterResponse, reqData *api.ClientRegisterRequest) (*api.QredoRegisterInitResponse, error) {
	reqDataInit := api.NewQredoRegisterInitRequest(reqData.Name, register.BLSPublicKey, register.ECPublicKey)
	return h.core.ClientInit(reqDataInit, register.RefID, reqData.APIKey, reqData.Base64PrivateKey)
}

func (h *SigningAgentHandler) finishRegistration(initResults *api.QredoRegisterInitResponse, refId string) error {
	reqDataFinish := &api.ClientRegisterFinishRequest{}

	// initResults contains only one extra field, timestamp
	if err := copier.Copy(&reqDataFinish, &initResults); err != nil {
		return err
	}

	if _, err := h.core.ClientRegisterFinish(reqDataFinish, refId); err != nil {
		h.log.Debugf("error while finishing client registration, %v", err)
		return err
	}

	return nil
}
