package hub

import (
	"sync"

	"go.uber.org/zap"

	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub/message"
)

// FeedHub maintains the set of active clients
// It provides ways to register and unregister clients
// Broadcasts messages from the source to all active clients
type FeedHub interface {
	Run() bool
	Stop()
	RegisterClient(client *FeedClient)
	UnregisterClient(client *FeedClient)
	IsRunning() bool
	ConnectedClients
}

// ConnectedClients gives access to the number of connected external clients of the hub
type ConnectedClients interface {
	GetExternalFeedClients() int
}

type feedHubImpl struct {
	source    Source
	broadcast chan []byte
	clients   map[*FeedClient]bool

	register   chan *FeedClient
	unregister chan *FeedClient
	log        *zap.SugaredLogger
	lock       sync.RWMutex
	isRunning  bool

	messageCache message.Cache
}

// NewFeedHub returns a FeedHub object that's an instance of FeedHubImpl
func NewFeedHub(source Source, log *zap.SugaredLogger, messageCache message.Cache) FeedHub {
	return &feedHubImpl{
		source:       source,
		log:          log,
		clients:      make(map[*FeedClient]bool),
		register:     make(chan *FeedClient),
		unregister:   make(chan *FeedClient),
		lock:         sync.RWMutex{},
		messageCache: messageCache,
	}
}

// IsRunning returns true only if the underlying source connection is open
func (w *feedHubImpl) IsRunning() bool {
	return w.isRunning
}

// Run makes sure the source is connected and the broadcast channel is ready to receive messages
func (w *feedHubImpl) Run() bool {
	if !w.source.Connect() {
		return false
	}
	var wg sync.WaitGroup
	wg.Add(2)

	//channel used to receive messages from the connection with the qredo server and send to all listening feed clients
	w.broadcast = w.source.GetSendChannel()

	go w.startHub(&wg)
	go w.source.Listen(&wg)

	wg.Wait() //wait for the hub to properly start and the source to start listening for messages
	return true
}

// Stop is closing the source connection
func (w *feedHubImpl) Stop() {
	if w.source.GetReadyState() == defs.ConnectionState.Open {
		w.source.Disconnect()
	}
}

// RegisterClient is adding a new active client to send messages to
func (w *feedHubImpl) RegisterClient(client *FeedClient) {
	w.lock.Lock()
	defer w.lock.Unlock()

	//send all previously received pending messages
	if w.messageCache != nil {
		messages := w.messageCache.GetMessages()
		for _, message := range messages {
			client.Feed <- message
		}
	}

	w.clients[client] = true
	w.log.Info("FeedHub: new client registered")
}

// UnregisterClient is removing a registered client and closes its Feed channel
func (w *feedHubImpl) UnregisterClient(client *FeedClient) {
	w.lock.Lock()
	defer w.lock.Unlock()

	registered := w.clients[client]
	if registered {
		close(client.Feed)
		delete(w.clients, client)
		w.log.Info("FeedHub: client unregistered")
	}
}

func (w *feedHubImpl) GetExternalFeedClients() int {
	w.lock.Lock()
	defer w.lock.Unlock()

	count := 0
	for fc := range w.clients {
		if !fc.IsInternal {
			count++
		}
	}

	return count
}

func (w *feedHubImpl) cleanUp() {
	w.lock.Lock()
	defer w.lock.Unlock()

	for client := range w.clients {
		w.log.Infof("FeedHub: closing feed clients")
		close(client.Feed)
		delete(w.clients, client)
	}
}

func (w *feedHubImpl) startHub(wg *sync.WaitGroup) {
	defer func() {
		w.isRunning = false
		w.cleanUp()
	}()

	w.isRunning = true
	wg.Done()

	for {
		if message, ok := <-w.broadcast; !ok {
			w.log.Infof("FeedHub: the broadcast channel was closed")
			return
		} else {
			w.lock.Lock()
			w.log.Infof("FeedHub: message received [%v]", string(message))

			if w.messageCache != nil {
				w.messageCache.AddMessage(message)
			}

			//send the message to all connected clients
			for client := range w.clients {
				client.Feed <- message
			}
			w.lock.Unlock()
		}
	}
}
