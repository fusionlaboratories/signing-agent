package clients

import (
	"net/http"
	"sync"

	"github.com/qredo/signing-agent/autoapprover"
	"github.com/qredo/signing-agent/clientfeed"
	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/lib"
	"go.uber.org/zap"
)

type AgentMng interface {
	RegisterClientFeed(w http.ResponseWriter, r *http.Request)
	ServiceMng
}

type ServiceMng interface {
	Start()
	Stop()
}

func NewManager(core lib.SigningAgentClient, feedHub hub.FeedHub, log *zap.SugaredLogger, config *config.Config, upgrader hub.WebsocketUpgrader, syncronizer autoapprover.ActionSyncronizer) AgentMng {
	return &clientsManager{
		feedHub: feedHub,
		core:    core,
		log:     log,
		config:  config,

		upgrader:          upgrader,
		newClientFeedFunc: clientfeed.NewClientFeed,
		syncronizer:       syncronizer,
	}
}

type newClientFeedFunc func(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister clientfeed.UnregisterFunc, config *config.WebSocketConfig) clientfeed.ClientFeed

type clientsManager struct {
	core              lib.SigningAgentClient
	feedHub           hub.FeedHub
	log               *zap.SugaredLogger
	config            *config.Config
	upgrader          hub.WebsocketUpgrader
	newClientFeedFunc newClientFeedFunc //function used by the feed clients to unregister themselves from the hub and stop receiving data
	syncronizer       autoapprover.ActionSyncronizer
}

// Start is running the feed hub if the agent is registered.
// It also makes sure the auto approver is registered to the hub and is listening for incoming actions, if enabled in the config
func (m *clientsManager) Start() {
	agentID := m.core.GetSystemAgentID()
	if len(agentID) == 0 {
		m.log.Info("Agent is not yet configured, auto-approval not started")
		return
	}

	autoApprover := m.newAutoApprover()

	if !m.feedHub.Run() {
		m.log.Error("failed to start the feed hub")
		if autoApprover != nil {
			autoApprover.Stop()
		}
		return
	}

	//feed hub is running, register the autoApprover if enabled
	if autoApprover != nil {
		m.feedHub.RegisterClient(&autoApprover.FeedClient)
	}
}

// Stop is called to stop the feed hub on request, by ex: when the service is stopped
func (m *clientsManager) Stop() {
	m.feedHub.Stop()
	m.log.Info("feed hub stopped")
}

func (m *clientsManager) RegisterClientFeed(w http.ResponseWriter, r *http.Request) {
	hubRunning := m.feedHub.IsRunning()

	if hubRunning {
		clientFeed := m.newClientFeed(w, r)
		if clientFeed != nil {
			var wg sync.WaitGroup
			wg.Add(1)

			go clientFeed.Start(&wg)
			wg.Wait() //wait for the client to set up the conn handling

			m.feedHub.RegisterClient(clientFeed.GetFeedClient())
			go clientFeed.Listen()
			m.log.Info("handler: connected to feed, listening ...")
		}
	} else {
		m.log.Debugf("handler: failed to connect, hub not running")
	}
}

func (m *clientsManager) newAutoApprover() *autoapprover.AutoApprover {
	if !m.config.AutoApprove.Enabled {
		m.log.Debug("Auto-approval feature not enabled in config")
		return nil
	}

	m.log.Debug("Auto-approval feature enabled")
	autoApprover := autoapprover.NewAutoApprover(m.core, m.log, m.config, m.syncronizer)

	var wg sync.WaitGroup
	wg.Add(1)

	go autoApprover.Listen(&wg)
	wg.Wait()

	return autoApprover
}

func (m *clientsManager) newClientFeed(w http.ResponseWriter, r *http.Request) clientfeed.ClientFeed {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Errorf("handler: failed to upgrade connection, err: %v", err)
		return nil
	}

	return m.newClientFeedFunc(conn, m.log, m.feedHub.UnregisterClient, &m.config.Websocket)
}
