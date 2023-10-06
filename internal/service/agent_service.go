package service

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/qredo/signing-agent/internal/action"
	"github.com/qredo/signing-agent/internal/api"
	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/autoapprover"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/feed"
	"github.com/qredo/signing-agent/internal/hub"

	"github.com/qredo/signing-agent/internal/store"
	"github.com/qredo/signing-agent/internal/util"
	"go.uber.org/zap"
)

type saveKeyDataRequest struct {
	BlsPublicKey string `json:"blsPublicKey"`
	EcPublicKey  string `json:"ecPublicKey"`
}

type getActionResponse struct {
	ActionID string `json:"actionID"`
}

type apiKeyNameResponse struct {
	Name string `json:"name"`
}

type AgentService interface {
	Start() error
	Stop()

	GetAgentDetails() (*api.GetAgentDetailsResponse, error)
	RegisterAgent(req *api.AgentRegisterRequest) (*api.AgentRegisterResponse, error)
	RegisterClientFeed(w http.ResponseWriter, r *http.Request)
	GetWebsocketStatus() *api.HealthCheckStatusResponse
}

type newClientFeedFunc func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister feed.UnregisterFunc, config config.WebSocketConfig) feed.ClientFeed
type genKeysFunc func() (string, string, string, string, error)

func NewAgentService(config config.Config, authProvider auth.HeaderProvider, store store.StoreWriter, signer action.Signer,
	feedHub hub.FeedHub, aa autoapprover.AutoApprover, log *zap.SugaredLogger, upgrader hub.WebsocketUpgrader,
	agentInfo *store.AgentInfo) AgentService {
	return &agentSrv{
		htc:               util.NewHTTPClient(),
		store:             store,
		config:            config,
		authProvider:      authProvider,
		signer:            signer,
		feedHub:           feedHub,
		autoApprover:      aa,
		log:               log,
		upgrader:          upgrader,
		newClientFeedFunc: feed.NewClientFeed,
		agentInfo:         agentInfo,
		genKeysFunc:       defs.GenerateKeys,
	}
}

type agentSrv struct {
	htc               *util.Client
	store             store.StoreWriter
	config            config.Config
	authProvider      auth.HeaderProvider
	signer            action.Signer
	feedHub           hub.FeedHub
	log               *zap.SugaredLogger
	upgrader          hub.WebsocketUpgrader
	newClientFeedFunc newClientFeedFunc //function used by the feed clients to unregister themselves from the hub and stop receiving data
	autoApprover      autoapprover.AutoApprover
	agentInfo         *store.AgentInfo
	genKeysFunc       genKeysFunc
}

// Start is running the feed hub if the agent is registered.
// It also makes sure the auto approver is registered to the hub and is listening for incoming actions, if enabled in the config
func (a *agentSrv) Start() error {
	if a.agentInfo == nil {
		a.log.Warn("Agent Service: agent is not yet configured, auto-approval not started")
		return nil
	}

	if a.autoApprover != nil {
		var wg sync.WaitGroup
		wg.Add(1)

		go a.autoApprover.Listen(&wg)
		wg.Wait()
	}

	if !a.feedHub.Run() {
		a.log.Error("Agent Service: failed to start the feed hub")
		if a.autoApprover != nil {
			a.autoApprover.Stop()
		}
		return fmt.Errorf("failed to start the feed hub")
	}

	//feed hub is running, register the autoApprover if enabled
	if a.autoApprover != nil {
		a.feedHub.RegisterClient(a.autoApprover.GetFeedClient())
	}

	return nil
}

// Stop is called to stop the service on request, by ex: when the app is stopped
func (a *agentSrv) Stop() {
	a.log.Info("Agent Service: stopping")

	a.feedHub.Stop()
	a.authProvider.Stop()
}

func (a *agentSrv) RegisterAgent(req *api.AgentRegisterRequest) (*api.AgentRegisterResponse, error) {
	if a.agentInfo != nil {
		return nil, defs.ErrBadRequest().WithDetail("signing agent already registered")
	}

	blsKeyPub, blsKeyPriv, ecKeyPub, ecKeyPriv, err := a.genKeysFunc()
	if err != nil {
		a.log.Errorf("Agent Service: error while generating keys, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to generate keys")
	}

	if err := a.authProvider.Initiate(req.WorkspaceID, req.Secret, req.APIKeyID); err != nil {
		return nil, defs.ErrInternal().WithDetail("failed to initiate the auth provider")
	}

	actionID, err := a.updateAPIKey(req.APIKeyID, req.WorkspaceID, blsKeyPub, ecKeyPub)
	if err != nil {
		a.log.Errorf("Agent Service: failed to update api keys, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to register agent")
	}

	if err := a.signer.SetKey(blsKeyPriv); err != nil {
		a.log.Errorf("Agent Service: failed to set signer key, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to setup signer")
	}

	if err = a.signer.ActionApprove(actionID); err != nil {
		return nil, err
	}

	a.agentInfo = &store.AgentInfo{
		BLSPrivateKey: blsKeyPriv,
		ECPrivateKey:  ecKeyPriv,
		WorkspaceID:   req.WorkspaceID,
		APIKeyID:      req.APIKeyID,
		APIKeySecret:  req.Secret,
	}

	if err = a.store.SaveAgentInfo(req.APIKeyID, a.agentInfo); err != nil {
		a.log.Errorf("Agent Service: failed to save agent info, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to save agent info")
	}

	name, err := a.getAgentName()
	if err != nil {
		a.log.Errorf("Agent Service: failed to get api key name, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to get agent name")
	}

	return &api.AgentRegisterResponse{
		GetAgentDetailsResponse: api.GetAgentDetailsResponse{
			Name:    name,
			AgentID: req.APIKeyID,
			FeedURL: a.getLocalFeed(),
		},
	}, nil
}

func (h agentSrv) GetWebsocketStatus() *api.HealthCheckStatusResponse {
	ws := h.feedHub.GetWebsocketStatus()

	resp := &api.HealthCheckStatusResponse{
		WebsocketStatus: ws,
		LocalFeedUrl:    h.getLocalFeed(),
	}

	return resp
}

func (a agentSrv) GetAgentDetails() (*api.GetAgentDetailsResponse, error) {
	if a.agentInfo == nil {
		return nil, defs.ErrBadRequest().WithDetail("agent not registered")
	}

	name, err := a.getAgentName()
	if err != nil {
		a.log.Errorf("Agent Service: failed to get api key name, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to get agent name")
	}

	return &api.GetAgentDetailsResponse{
		Name:    name,
		AgentID: a.agentInfo.APIKeyID,
		FeedURL: a.getLocalFeed(),
	}, nil
}

func (a *agentSrv) RegisterClientFeed(w http.ResponseWriter, r *http.Request) {
	if !a.feedHub.IsRunning() {
		a.log.Errorf("Agent Service: failed to connect feed client, hub not running")
		return
	}

	clientFeed := a.newClientFeed(w, r)
	if clientFeed != nil {
		var wg sync.WaitGroup
		wg.Add(2)

		go clientFeed.Start(&wg)
		go clientFeed.Listen(&wg)

		wg.Wait() //wait for the client to set up the conn handling and start listening

		a.feedHub.RegisterClient(clientFeed.GetFeedClient())
		a.log.Info("Agent Service: new local feed client connected")
	}
}

func (a agentSrv) updateAPIKey(APIkeyID, workspaceID, blsKey, ecKey string) (string, error) {
	req := saveKeyDataRequest{
		BlsPublicKey: blsKey,
		EcPublicKey:  ecKey,
	}

	resp := &getActionResponse{}
	header := a.authProvider.GetAuthHeader()

	if err := a.htc.Request(http.MethodPut, defs.URLAPIKey(a.config.Base.QredoAPI, workspaceID, APIkeyID), req, resp, header); err != nil {
		return defs.EmptyString, err
	}

	return resp.ActionID, nil
}

func (a agentSrv) getAgentName() (string, error) {
	resp := &apiKeyNameResponse{}
	header := a.authProvider.GetAuthHeader()
	url := defs.URLAPIKey(a.config.Base.QredoAPI, a.agentInfo.WorkspaceID, a.agentInfo.APIKeyID)

	if err := a.htc.Request(http.MethodGet, url, nil, resp, header); err != nil {
		return defs.EmptyString, err
	}

	return resp.Name, nil
}

func (a agentSrv) newClientFeed(w http.ResponseWriter, r *http.Request) feed.ClientFeed {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		a.log.Errorf("Agent Service: failed to upgrade connection, err: %v", err)
		return nil
	}

	return a.newClientFeedFunc(conn, a.log, a.feedHub.UnregisterClient, a.config.Websocket)
}

func (a agentSrv) getLocalFeed() string {
	return defs.URLlocalFeed(a.config.HTTP.Addr)
}
