package clientfeed

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/hub"
	"github.com/qredo/signing-agent/util"
)

func TestClientFeedImpl_GetFeedClient(t *testing.T) {
	//Arrange
	sut := NewClientFeed(nil, nil, nil, &config.WebSocketConfig{})

	//Act
	res := sut.GetFeedClient()

	//Assert
	assert.NotNil(t, res)
	assert.False(t, res.IsInternal)
}

func TestClientFeedImpl_Start_unregisters_the_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockConn := &hub.MockWebsocketConnection{
		NextError: errors.New("some write control error"),
	}
	var lastUnregisteredClient *hub.FeedClient
	unregister := func(client *hub.FeedClient) {
		lastUnregisteredClient = client
	}
	sut := NewClientFeed(mockConn, util.NewTestLogger(), unregister, &config.WebSocketConfig{
		PingPeriod: 2,
		PongWait:   2,
		WriteWait:  2,
	})
	var wg sync.WaitGroup
	wg.Add(1)

	//Act
	sut.Start(&wg)
	wg.Wait()

	//Assert
	assert.True(t, mockConn.SetPongHandlerCalled)
	assert.True(t, mockConn.SetPingHandlerCalled)
	assert.True(t, mockConn.WriteControlCalled)
	assert.True(t, mockConn.CloseCalled)
	assert.Equal(t, websocket.PingMessage, mockConn.LastMessageType)
	assert.Empty(t, mockConn.LastData)
	assert.Equal(t, sut.GetFeedClient(), lastUnregisteredClient)
}

func TestClientFeedImpl_Listen_writes_the_message(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockConn := &hub.MockWebsocketConnection{
		NextError: errors.New("some write error"),
	}

	sut := &clientFeedImpl{
		conn:       mockConn,
		log:        util.NewTestLogger(),
		readyState: defs.ConnectionState.Closed,
		FeedClient: hub.NewFeedClient(false),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	//Act

	go sut.Listen(&wg)
	wg.Wait()
	sut.Feed <- []byte("some message")
	<-time.After(time.Second) //give it time to process the message

	//Assert
	assert.True(t, mockConn.WriteMessageCalled)
	assert.Equal(t, websocket.TextMessage, mockConn.LastMessageType)
	assert.Equal(t, "some message", string(mockConn.LastData))
	close(sut.Feed)
}

func TestClientFeedImpl_Listen_closes_connection(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t)
	mockConn := &hub.MockWebsocketConnection{}

	sut := &clientFeedImpl{
		conn:       mockConn,
		log:        util.NewTestLogger(),
		readyState: defs.ConnectionState.Open,
		FeedClient: hub.NewFeedClient(false),
		writeWait:  2,
		pingPeriod: 2,
		pongWait:   2,
		closeConn:  make(chan bool),
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go sut.Start(&wg)

	//Act
	go sut.Listen(&wg)
	wg.Wait()
	close(sut.Feed)
	<-time.After(time.Second) //give it time to processs

	//Assert
	assert.True(t, mockConn.WriteControlCalled)
	assert.Equal(t, websocket.CloseMessage, mockConn.LastMessageType)
	assert.Equal(t, defs.ConnectionState.Closed, sut.readyState)
}
