package hub

import (
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/lib"
)

// Source has an underlying websocket connection used to receive messages. It will send these messages to an outbound channel
type Source interface {
	Connect() bool
	Disconnect()
	Listen(wg *sync.WaitGroup)
	GetSendChannel() chan []byte
	SourceStats
}

// SourceStats gives access to the feed url and the connection's ready state
type SourceStats interface {
	GetFeedUrl() string
	GetReadyState() string
}

type websocketSource struct {
	dialer               WebsocketDialer
	conn                 WebsocketConnection
	log                  *zap.SugaredLogger
	core                 lib.SigningAgentClient
	feedUrl              string
	shouldReconnect      bool
	readyState           string
	reconnectIntervalMax time.Duration
	reconnectInterval    time.Duration
	rxMessages           chan []byte
	lock                 sync.RWMutex
}

// NewWebsocketSource returns a Source object that's an instance of websocketSource
func NewWebsocketSource(dialer WebsocketDialer, feedUrl string, log *zap.SugaredLogger, core lib.SigningAgentClient, config *config.WebSocketConfig) Source {
	return &websocketSource{
		log:                  log,
		core:                 core,
		feedUrl:              feedUrl,
		dialer:               dialer,
		shouldReconnect:      true,
		readyState:           defs.ConnectionState.Closed,
		reconnectIntervalMax: time.Duration(config.ReconnectTimeOut) * time.Second,
		reconnectInterval:    time.Duration(config.ReconnectInterval) * time.Second,
		rxMessages:           make(chan []byte),
		lock:                 sync.RWMutex{},
	}
}

// Connect is trying to establish a websocket connection which will be used as a source
// It tries to reconnect at each interval defined in the configuration
func (w *websocketSource) Connect() bool {
	w.setReadyState(defs.ConnectionState.Connecting)

	startTime := time.Now()
	for time.Since(startTime) < w.reconnectIntervalMax {
		if err := w.dial(); err == nil {
			w.log.Infof("WebsocketSource: connected to feed %v", w.feedUrl)
			return true
		} else {
			w.log.Errorf("WebsocketSource: cannot connect to feed: %v, retry connection in %v", err, w.reconnectInterval)
			time.Sleep(w.reconnectInterval)
		}
	}

	w.setReadyState(defs.ConnectionState.Closed)
	return false
}

// Listen is receiving messages from the underlying websocket connection and sends them to the outbound channel
// In case of a reading or connectivity issue it tries to reconnect
func (w *websocketSource) Listen(wg *sync.WaitGroup) {
	defer func() {
		w.conn.Close()
		close(w.rxMessages)
	}()

	wg.Done()

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			//closed on request
			if !w.shouldReconnect {
				return
			}
			//either connection issue or issue reading the message
			w.log.Errorf("WebsocketSource: unexpected connection error: %v", err)
			if !w.Connect() {
				return
			}
		} else {
			w.rxMessages <- message
		}
	}

}

// Disconnect is closing the websocket upon request and signals the reconnect should not happen
func (w *websocketSource) Disconnect() {
	w.log.Infof("WebsocketSource: disconnecting from feed %v", w.feedUrl)
	if err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
		w.log.Errorf("WebsocketSource: error on send CloseMessage, error: %v", err)
	}
	w.shouldReconnect = false
	w.setReadyState(defs.ConnectionState.Closed)
}

// GetFeedUrl returns the websocket url
func (w *websocketSource) GetFeedUrl() string {
	return w.feedUrl
}

// GetReadyState returns the status of the websocket connection
func (w *websocketSource) GetReadyState() string {
	w.lock.RLock()
	defer w.lock.RUnlock()

	return w.readyState
}

// GetSendChannel returns the outbound channel
func (w *websocketSource) GetSendChannel() chan []byte {
	return w.rxMessages
}

func (w *websocketSource) dial() error {
	headers, err := w.genConnectHeaders()
	if err != nil {
		w.log.Errorf("WebsocketSource: failed to generate dial headers: %v", err)
		return err
	}

	conn, _, err := w.dialer.Dial(w.feedUrl, headers)
	if err == nil {
		w.conn = conn
		w.setReadyState(defs.ConnectionState.Open)
		return nil
	}

	return err
}

func (w *websocketSource) genConnectHeaders() (http.Header, error) {
	zkpOnePass, err := w.core.GetAgentZKPOnePass()
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set(defs.AuthHeader, hex.EncodeToString(zkpOnePass))
	return headers, nil
}

func (w *websocketSource) setReadyState(state string) {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.readyState = state
}
