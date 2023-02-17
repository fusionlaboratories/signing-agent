package lib

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/defs"
)

const WsTimeout = time.Second * 60

type Handler func(message []byte)

type ErrHandler func(err error)

type ServeCB func(string, Handler, ErrHandler) (chan struct{}, chan struct{}, error)

type WsActionInfoEvent struct {
	ID         string `json:"id"`
	AgentID    string `json:"coreClientID"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Timestamp  int64  `json:"timestamp"`
	ExpireTime int64  `json:"expireTime"`
}

type WsActionInfoHandler func(event *WsActionInfoEvent)

func NewFeed(feedUrl string, agent SigningAgentClient, serveFu ServeCB) *feed {
	fd := &feed{
		feedUrl: feedUrl,
		agent:   agent,
	}

	// sets default Serve function if not provided
	fd.servFu = serveFu
	if fd.servFu == nil {
		fd.servFu = fd.Serve
	}
	return fd
}

type feed struct {
	agent   SigningAgentClient
	feedUrl string
	servFu  ServeCB
}

func (f *feed) ActionEvent(handler WsActionInfoHandler, errHandler ErrHandler) (doneCH, stopCH chan struct{}, err error) {
	wsHandler := func(message []byte) {
		event := &WsActionInfoEvent{}
		if err := json.Unmarshal(message, &event); err != nil {
			errHandler(err)
			return
		}
		handler(event)
	}
	return f.servFu(f.feedUrl, wsHandler, errHandler)
}

func (f *feed) keepAlive(con *websocket.Conn, timeout time.Duration) {
	tk := time.NewTicker(timeout)
	lastResponse := time.Now()
	con.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer tk.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			if err := con.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
				return
			}
			<-tk.C
			if time.Since(lastResponse) > timeout {
				_ = con.Close()
				return
			}
		}
	}()
}

func (f *feed) Serve(url string, handler Handler, errHandler ErrHandler) (doneCH, stopCH chan struct{}, err error) {
	if len(f.agent.GetSystemAgentID()) == 0 {
		return nil, nil, errors.New("cannot get agent-id")
	}

	zkpToken, err := f.agent.GetAgentZKPOnePass()
	if err != nil {
		return nil, nil, errors.Wrap(err, "get zkp token")
	}

	headers := http.Header{}
	headers.Set(defs.AuthHeader, hex.EncodeToString(zkpToken))

	c, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, nil, err
	}

	doneCH = make(chan struct{})
	stopCH = make(chan struct{})

	go func() {
		defer close(doneCH)
		f.keepAlive(c, WsTimeout)

		go func() {
			select {
			case <-stopCH:
			case <-doneCH:
			}
			_ = c.Close()
		}()

		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				errHandler(err)
				return
			}
			handler(msg)
		}
	}()

	return
}

func (h *signingAgent) ReadAction(feedUrl string, cb ServeCB) *feed {
	return NewFeed(feedUrl, h, cb)
}
