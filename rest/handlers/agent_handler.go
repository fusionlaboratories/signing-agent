package handlers

import (
	"net/http"
	"sync"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/api"
	"github.com/qredo/signing-agent/autoapprover"
	"github.com/qredo/signing-agent/clientfeed"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

type newClientFeedFunc func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister clientfeed.UnregisterFunc, config *config.WebSocketConfig) clientfeed.ClientFeed

type SigningAgentHandler struct {
	feedHub           hub.FeedHub
	log               *zap.SugaredLogger
	core              lib.SigningAgentClient
	config            *config.AutoApprove
	websocketConfig   *config.WebSocketConfig
	localFeed         string
	decode            func(interface{}, *http.Request) error
	autoApprover      *autoapprover.AutoApprover
	upgrader          hub.WebsocketUpgrader
	newClientFeedFunc newClientFeedFunc //function used by the feed clients to unregister themselves from the hub and stop receiving data
}

// NewSigningAgentHandler instantiates and returns a new SigningAgentHandler object.
func NewSigningAgentHandler(feedHub hub.FeedHub, core lib.SigningAgentClient, log *zap.SugaredLogger, config *config.Config, autoApprover *autoapprover.AutoApprover, upgrader hub.WebsocketUpgrader, localFeed string) *SigningAgentHandler {
	return &SigningAgentHandler{
		feedHub:           feedHub,
		log:               log,
		core:              core,
		config:            &config.AutoApprove,
		localFeed:         localFeed,
		decode:            util.DecodeRequest,
		autoApprover:      autoApprover,
		upgrader:          upgrader,
		websocketConfig:   &config.Websocket,
		newClientFeedFunc: clientfeed.NewClientFeed,
	}
}

// StartAgent is running the feed hub if the agent is registered.
// It also makes sure the auto approver is registered to the hub and is listening for incoming actions, if enabled in the config
func (h *SigningAgentHandler) StartAgent() {
	agentID := h.core.GetSystemAgentID()
	if len(agentID) == 0 {
		h.log.Info("Agent is not yet configured, auto-approval not started")
		return
	}

	if !h.feedHub.Run() {
		h.log.Error("failed to start the feed hub")
		return
	}

	if !h.config.Enabled {
		h.log.Debug("Auto-approval feature not enabled in config")
		return
	}

	h.feedHub.RegisterClient(&h.autoApprover.FeedClient)
	go h.autoApprover.Listen()
}

// StopAgent is called to stop the feed hub on request, by ex: when the service is stopped
func (h *SigningAgentHandler) StopAgent() {
	h.feedHub.Stop()
	h.log.Info("feed hub stopped")
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
		h.StartAgent()
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
	hubRunning := h.feedHub.IsRunning()

	if hubRunning {
		clientFeed := h.newClientFeed(w, r)
		if clientFeed != nil {
			var wg sync.WaitGroup
			wg.Add(1)

			go clientFeed.Start(&wg)
			wg.Wait() //wait for the client to set up the conn handling

			h.feedHub.RegisterClient(clientFeed.GetFeedClient())
			go clientFeed.Listen()
			h.log.Info("handler: connected to feed, listening ...")
		}
	} else {
		h.log.Debugf("handler: failed to connect, hub not running")
	}

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

func (h *SigningAgentHandler) newClientFeed(w http.ResponseWriter, r *http.Request) clientfeed.ClientFeed {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorf("handler: failed to upgrade connection, err: %v", err)
		return nil
	}

	return h.newClientFeedFunc(conn, h.log, h.feedHub.UnregisterClient, h.websocketConfig)
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
