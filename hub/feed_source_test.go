package hub

import (
	"sync"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/lib"
	"github.com/qredo/signing-agent/util"
)

func TestWebsocketSource_Connect_retries_on_fail_to_get__ZKPOnePass(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)

	mock_dialer := &mockWebsocketDialer{}
	mock_core := &lib.MockSigningAgentClient{
		NextError: errors.New("some error"),
	}
	sut := NewWebsocketSource(mock_dialer, "feed", util.NewTestLogger(), mock_core, &config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2})

	//Act
	res := sut.Connect()

	//Assert
	assert.False(t, res)
	assert.True(t, mock_core.GetAgentZKPOnePassCalled)
	assert.False(t, mock_dialer.DialCalled)
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.Equal(t, 3, mock_core.Counter)
}

func TestWebsocketSource_Connects(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)

	mock_dialer := &mockWebsocketDialer{
		NextConn: &websocket.Conn{},
	}
	mock_core := &lib.MockSigningAgentClient{
		NextZKPOnePass: []byte("zkp"),
	}
	sut := NewWebsocketSource(mock_dialer, "feed", util.NewTestLogger(), mock_core, &config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2})

	//Act
	res := sut.Connect()

	//Assert
	assert.True(t, res)
	assert.True(t, mock_core.GetAgentZKPOnePassCalled)
	assert.Equal(t, 1, mock_core.Counter)
	assert.True(t, mock_dialer.DialCalled)
	assert.Equal(t, "feed", mock_dialer.LastFeedUrl)
	assert.Equal(t, "7a6b70", mock_dialer.LastRequestHeader.Get("X-Api-Zkp"))
	assert.Equal(t, defs.ConnectionState.Open, sut.GetReadyState())
}

func TestWebsocketSource_retries_on_dial_error(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_dialer := &mockWebsocketDialer{
		NextConn:  &websocket.Conn{},
		NextError: errors.New("some error"),
	}
	mock_core := &lib.MockSigningAgentClient{
		NextZKPOnePass: []byte("zkp"),
	}
	conn := NewWebsocketSource(mock_dialer, "feed", util.NewTestLogger(), mock_core, &config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2})

	//Act
	res := conn.Connect()

	//Assert
	assert.False(t, res)
	assert.True(t, mock_dialer.DialCalled)
	assert.Equal(t, "feed", mock_dialer.LastFeedUrl)
	assert.Equal(t, "7a6b70", mock_dialer.LastRequestHeader.Get("X-Api-Zkp"))
	assert.True(t, mock_core.GetAgentZKPOnePassCalled)
	assert.Equal(t, 3, mock_core.Counter)
	assert.Equal(t, defs.ConnectionState.Closed, conn.GetReadyState())
}

func TestWebsocketSource_GetFeedUrl(t *testing.T) {
	//Arrange
	sut := NewWebsocketSource(nil, "feed", nil, nil, &config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2})

	//Act
	res := sut.GetFeedUrl()

	//Assert
	assert.Equal(t, "feed", res)
}

func TestWebsocketSource_Disconnect(t *testing.T) {
	//Arrange
	mock_conn := &MockWebsocketConnection{
		NextError: errors.New("some error"),
	}
	sut := &websocketSource{
		conn:            mock_conn,
		shouldReconnect: true,
		readyState:      defs.ConnectionState.Open,
		log:             util.NewTestLogger(),
	}

	//Act
	sut.Disconnect()

	//Assert
	assert.False(t, sut.shouldReconnect)
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, mock_conn.WriteMessageCalled)
	assert.Equal(t, websocket.CloseMessage, mock_conn.LastMessageType)
}

func TestWebsocketSource_Listen_shouldnt_reconnect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_conn := &MockWebsocketConnection{
		NextError: errors.New("some error"),
		read:      make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:            mock_conn,
		shouldReconnect: false,
		readyState:      defs.ConnectionState.Closed,
		log:             util.NewTestLogger(),
		rxMessages:      make(chan []byte),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	//Act
	go sut.Listen(&wg)
	wg.Wait()
	mock_conn.read <- true

	//Assert
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, mock_conn.ReadMessageCalled)

	_, ok := <-sut.rxMessages //channel was closed
	assert.False(t, ok)
}

func TestWebsocketSource_Listen_don_t_reconnect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_conn := &MockWebsocketConnection{
		NextError: errors.New("some error"),
		read:      make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:            mock_conn,
		shouldReconnect: true,
		readyState:      defs.ConnectionState.Closed,
		log:             util.NewTestLogger(),
		rxMessages:      make(chan []byte),
	}

	var wg sync.WaitGroup
	wg.Add(1)

	//Act
	go sut.Listen(&wg)
	wg.Wait()
	mock_conn.read <- true

	//Assert
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, mock_conn.ReadMessageCalled)

	_, ok := <-sut.rxMessages //channel was closed
	assert.False(t, ok)
}

func TestWebsocketSource_Listen_sends_message(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mock_conn := &MockWebsocketConnection{
		NextMessageType: 1,
		NextData:        []byte("some message"),
		read:            make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:       mock_conn,
		rxMessages: make(chan []byte),
	}
	var (
		message       []byte
		chanOk        bool
		wg, wg_client sync.WaitGroup
	)

	wg.Add(1)

	//Act
	go sut.Listen(&wg)

	wg.Wait()
	wg_client.Add(1)

	go func() {
		msg, ok := <-sut.GetSendChannel()
		message = msg
		chanOk = ok
		wg_client.Done()
	}()

	mock_conn.read <- true
	wg_client.Wait()

	//Assert
	assert.True(t, mock_conn.ReadMessageCalled)
	assert.True(t, chanOk)
	assert.Equal(t, []byte("some message"), message)

	//Clean up
	mock_conn.NextError = errors.New("some error")
	sut.shouldReconnect = false
	mock_conn.read <- true
}
