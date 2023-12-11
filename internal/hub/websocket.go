package hub

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketConnection is a wrapper around the websocket connection
type WebsocketConnection interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	WriteControl(messageType int, data []byte, deadline time.Time) error
	SetReadDeadline(t time.Time) error
	SetPongHandler(h func(appData string) error)
	SetPingHandler(h func(appData string) error)
	SetWriteDeadline(t time.Time) error
	Close() error
}

// WebsocketDialer is a wrapper around the websocket dialer
type WebsocketDialer interface {
	Dial(url string, requestHeader http.Header) (WebsocketConnection, *http.Response, error)
}

type defaultDialer struct {
	dialer *websocket.Dialer
}

// NewDefaultDialer returns a new WebsocketDialer that's an instance of the websocket DefaultDialer
func NewDefaultDialer() WebsocketDialer {
	return &defaultDialer{
		dialer: websocket.DefaultDialer,
	}
}

// Dial establishes a new websocket connection
func (d *defaultDialer) Dial(url string, requestHeader http.Header) (WebsocketConnection, *http.Response, error) {
	return d.dialer.Dial(url, requestHeader)
}

// WebsocketUpgrader is a wrapper around the websocket Upgrader
type WebsocketUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketConnection, error)
}

type defaultUpgrader struct {
	upgrader *websocket.Upgrader
}

// NewDefaultUpgrader returns a new WebsocketUpgrader that's an instance of the websocket Upgrader
func NewDefaultUpgrader(readBufferSize, writeBufferSize int) WebsocketUpgrader {
	up := &defaultUpgrader{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
		},
	}
	return up
}

// Upgrade is establishing a websocket connection from an incoming request
func (d *defaultUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketConnection, error) {
	return d.upgrader.Upgrade(w, r, responseHeader)
}
