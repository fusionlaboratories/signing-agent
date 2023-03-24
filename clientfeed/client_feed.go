// Package clientfeed provides functionality to register to a feed hub to receive bytes data as TextMessage
// The received data is being then written also as a TextMessage to an open websocket connection.

package clientfeed

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
)

// UnregisterFunc is used by the ClientFeed to unregister itself from the feed hub. Upon its request, the Feed channel will be closed and no data will be received
type UnregisterFunc func(client *hub.FeedClient)

// ClientFeed is a client recieving messages from a feeb hub it's registered to
type ClientFeed interface {
	Start(wg *sync.WaitGroup)
	Listen(wg *sync.WaitGroup)
	GetFeedClient() *hub.FeedClient
}

type clientFeedImpl struct {
	hub.FeedClient
	conn       hub.WebsocketConnection
	log        *zap.SugaredLogger
	closeConn  chan bool
	pongWait   time.Duration
	writeWait  time.Duration
	pingPeriod time.Duration
	readyState string
	unregister UnregisterFunc
}

// NewClientFeed returns a new ClientFeed which is an instance of ClientFeedImpl initialized with the provided parameters
// ClientFeed has an external FeedClient which means it can unregister itself from the feed hub to stop receiving data
func NewClientFeed(conn hub.WebsocketConnection, log *zap.SugaredLogger, unregister UnregisterFunc, config *config.WebSocketConfig) ClientFeed {
	return &clientFeedImpl{
		FeedClient: hub.NewFeedClient(false),
		conn:       conn,
		log:        log,
		closeConn:  make(chan bool),
		writeWait:  time.Duration(config.WriteWait) * time.Second,
		pongWait:   time.Duration(config.PongWait) * time.Second,
		pingPeriod: time.Duration(config.PingPeriod) * time.Second,
		readyState: defs.ConnectionState.Open,
		unregister: unregister,
	}
}

// GetFeedClient returns the internal FeedClient structure used to register itself to the feed hub in order to start receiving data on the Feed channel
func (c *clientFeedImpl) GetFeedClient() *hub.FeedClient {
	return &c.FeedClient
}

// Start is setting up the handlers for maintaining the connection
// It is also responsible for closing the connection when signaled to do
// The wg is used to signal the caller of the function when the handlers are set and the client is prepared to function
func (c *clientFeedImpl) Start(wg *sync.WaitGroup) {
	c.log.Debug("ClientFeed - starting the client feed")

	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		c.conn.Close()
		ticker.Stop()
		c.readyState = defs.ConnectionState.Closed
	}()

	c.setHandlers()
	wg.Done() //notify the caller everything is properly set up and ready to receive/send messages

	for {
		select {
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(c.writeWait)) // result is always nil
			if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(c.pingPeriod)); err != nil {
				c.log.Errorf("ClientFeed - websocket PingMessage found broken pipe, terminating, err: %v", err)
				c.readyState = defs.ConnectionState.Closed

				//must also unregister from the feed hub to stop receiving messages
				c.unregister(&c.FeedClient)
				return
			}
		case <-c.closeConn:
			c.log.Debug("ClientFeed - closing websocket connection")
			if err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(c.writeWait)); err != nil {
				c.log.Errorf("ClientFeed - websocket CloseMessage err: %v", err)
			}
			return
		}
	}
}

// Listen is receiving data on the Feed channel and writes them to the opened websocket connection
func (c *clientFeedImpl) Listen(wg *sync.WaitGroup) {
	wg.Done()
	for {
		if message, ok := <-c.Feed; !ok {
			c.log.Debug("ClientFeed: client feed channel was closed")
			//channel was closed by the feed hub so we must close the websocket connection if still open
			if c.readyState == defs.ConnectionState.Open {
				c.closeConn <- true
			}
			return
		} else {
			c.log.Infof("ClientFeed: writing message to websocket conn: %v", string(message))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.log.Errorf("ClientFeed: error while writing data to websocket conn:%v", err)
			}
		}
	}
}

func (c *clientFeedImpl) setHandlers() {
	c.conn.SetPongHandler(func(message string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(c.pongWait)); err != nil {
			return err
		}

		return c.conn.WriteControl(websocket.PingMessage, []byte(message), time.Now().Add(c.writeWait))
	})

	c.conn.SetPingHandler(func(message string) error {
		if err := c.conn.SetWriteDeadline(time.Now().Add(c.pingPeriod)); err != nil {
			return err
		}

		return c.conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(c.writeWait))
	})
}
