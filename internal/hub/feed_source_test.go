package hub

import (
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/gorilla/websocket"
	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
)

type mockWebsocketDialer struct {
	DialCalled        bool
	LastFeedUrl       string
	LastRequestHeader http.Header
	NextError         error
	NextConn          WebsocketConnection
}

func (m *mockWebsocketDialer) Dial(url string, requestHeader http.Header) (WebsocketConnection, *http.Response, error) {
	m.DialCalled = true
	m.LastFeedUrl = url
	m.LastRequestHeader = requestHeader
	return m.NextConn, nil, m.NextError
}

func TestWebsocketSource_Connects(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	dialerMock := &mockWebsocketDialer{
		NextConn: &websocket.Conn{},
	}
	testHeader := http.Header{}
	testHeader.Set("x-token", "test values")
	authMock := &auth.MockHeaderProvider{
		NextHeader: testHeader,
	}
	sut := NewWebsocketSource(dialerMock, "feed", util.NewTestLogger(), config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2}, authMock)

	//Act
	res := sut.Connect()

	//Assert
	assert.True(t, res)
	assert.True(t, authMock.GetAuthHeaderCalled)
	assert.True(t, dialerMock.DialCalled)
	assert.Equal(t, "feed", dialerMock.LastFeedUrl)
	assert.Equal(t, "test values", dialerMock.LastRequestHeader.Get("x-token"))
	assert.Equal(t, defs.ConnectionState.Open, sut.GetReadyState())
}

func TestWebsocketSource_retries_on_dial_error(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	dialerMock := &mockWebsocketDialer{
		NextConn:  &websocket.Conn{},
		NextError: errors.New("some error"),
	}
	testHeader := http.Header{}
	testHeader.Set("x-token", "test values")
	authMock := &auth.MockHeaderProvider{
		NextHeader: testHeader,
	}
	sut := NewWebsocketSource(dialerMock, "feed", util.NewTestLogger(), config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2}, authMock)

	//Act
	res := sut.Connect()

	//Assert
	assert.False(t, res)
	assert.True(t, dialerMock.DialCalled)
	assert.Equal(t, "feed", dialerMock.LastFeedUrl)
	assert.Equal(t, "test values", dialerMock.LastRequestHeader.Get("x-token"))
	assert.True(t, authMock.GetAuthHeaderCalled)
	assert.Equal(t, 3, authMock.Counter)
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
}

func TestWebsocketSource_GetFeedUrl(t *testing.T) {
	//Arrange
	sut := NewWebsocketSource(nil, "feed", nil, config.WebSocketConfig{ReconnectTimeOut: 6, ReconnectInterval: 2}, nil)

	//Act
	res := sut.GetFeedUrl()

	//Assert
	assert.Equal(t, "feed", res)
}

func TestWebsocketSource_Disconnect(t *testing.T) {
	//Arrange
	connMock := &MockWebsocketConnection{
		NextError: errors.New("some error"),
	}
	sut := &websocketSource{
		conn:            connMock,
		shouldReconnect: true,
		readyState:      defs.ConnectionState.Open,
		log:             util.NewTestLogger(),
	}

	//Act
	sut.Disconnect()

	//Assert
	assert.False(t, sut.shouldReconnect)
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, connMock.WriteMessageCalled)
	assert.Equal(t, websocket.CloseMessage, connMock.LastMessageType)
}

func TestWebsocketSource_Listen_shouldnt_reconnect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	connMock := &MockWebsocketConnection{
		NextError: errors.New("some error"),
		read:      make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:            connMock,
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
	connMock.read <- true

	//Assert
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, connMock.ReadMessageCalled)

	_, ok := <-sut.rxMessages //channel was closed
	assert.False(t, ok)
}

func TestWebsocketSource_Listen_don_t_reconnect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	connMock := &MockWebsocketConnection{
		NextError: errors.New("some error"),
		read:      make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:            connMock,
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
	connMock.read <- true

	//Assert
	assert.Equal(t, defs.ConnectionState.Closed, sut.GetReadyState())
	assert.True(t, connMock.ReadMessageCalled)

	_, ok := <-sut.rxMessages //channel was closed
	assert.False(t, ok)
}

func TestWebsocketSource_Listen_sends_message(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	connMock := &MockWebsocketConnection{
		NextMessageType: 1,
		NextData:        []byte("some message"),
		read:            make(chan bool, 1),
	}

	sut := &websocketSource{
		conn:       connMock,
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

	connMock.read <- true
	wg_client.Wait()

	//Assert
	assert.True(t, connMock.ReadMessageCalled)
	assert.True(t, chanOk)
	assert.Equal(t, []byte("some message"), message)

	//Clean up
	connMock.NextError = errors.New("some error")
	sut.shouldReconnect = false
	connMock.read <- true
}
