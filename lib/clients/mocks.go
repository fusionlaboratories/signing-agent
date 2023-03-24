package clients

import (
	"sync"
	"time"

	"github.com/qredo/signing-agent/hub"
)

type mockFeedHub struct {
	NextRun                bool
	RunCalled              bool
	RegisterClientCalled   bool
	UnregisterClientCalled bool
	StopCalled             bool
	IsRunningCalled        bool
	LastRegisteredClient   *hub.FeedClient
	LastUnregisteredClient *hub.FeedClient
}

func (m *mockFeedHub) IsRunning() bool {
	m.IsRunningCalled = true
	return m.NextRun
}

func (m *mockFeedHub) Run() bool {
	m.RunCalled = true
	return m.NextRun
}

func (m *mockFeedHub) Stop() {
	m.StopCalled = true
}

func (m *mockFeedHub) RegisterClient(client *hub.FeedClient) {
	m.RegisterClientCalled = true
	m.LastRegisteredClient = client
}

func (m *mockFeedHub) UnregisterClient(client *hub.FeedClient) {
	m.UnregisterClientCalled = true
	m.LastUnregisteredClient = client
}

func (m *mockFeedHub) GetExternalFeedClients() int {
	return 0
}

type mockClientFeed struct {
	StartCalled         bool
	ListenCalled        bool
	GetFeedClientCalled bool
	NextFeedClient      *hub.FeedClient
}

func (m *mockClientFeed) Start(wg *sync.WaitGroup) {
	m.StartCalled = true
	//add a bit of delay to ensure the sync
	<-time.After(time.Second)
	wg.Done()
}

func (m *mockClientFeed) Listen(wg *sync.WaitGroup) {
	m.ListenCalled = true
	//add a bit of delay to ensure the sync
	<-time.After(2 * time.Second)
	wg.Done()
}

func (m *mockClientFeed) GetFeedClient() *hub.FeedClient {
	m.GetFeedClientCalled = true
	return m.NextFeedClient
}
