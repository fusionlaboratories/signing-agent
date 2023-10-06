package hub

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"

	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/hub/message"
	"github.com/qredo/signing-agent/internal/util"
)

type mockSourceConnection struct {
	ConnectCalled       bool
	ListenCalled        bool
	DisconnectCalled    bool
	GetReadyStateCalled bool
	NextConnect         bool
	NextReadyState      string
	RxMessages          chan []byte
	NextFeedURL         string
}

func (m *mockSourceConnection) Connect() bool {
	m.ConnectCalled = true
	return m.NextConnect
}

func (m *mockSourceConnection) Disconnect() {
	m.DisconnectCalled = true

}

func (m *mockSourceConnection) Listen(wg *sync.WaitGroup) {
	m.ListenCalled = true
	wg.Done()
}

func (m *mockSourceConnection) GetFeedUrl() string {
	return m.NextFeedURL
}

func (m *mockSourceConnection) GetReadyState() string {
	m.GetReadyStateCalled = true
	return m.NextReadyState
}

func (m *mockSourceConnection) GetSendChannel() chan []byte {
	return m.RxMessages
}

var ignoreOpenCensus = goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")

func TestFeedHub_Run_fails_to_connect(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockSourceConn := &mockSourceConnection{}
	feedHub := NewFeedHub(mockSourceConn, util.NewTestLogger(), nil)

	//Act
	res := feedHub.Run()

	//Assert
	assert.False(t, res)
	assert.True(t, mockSourceConn.ConnectCalled)
	assert.False(t, feedHub.IsRunning())
}

func TestFeedHub_Run_connects_and_listens(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockSourceConn := &mockSourceConnection{
		NextConnect: true,
		RxMessages:  make(chan []byte, 1),
	}
	mockCache := &message.MockCache{}
	feedHub := NewFeedHub(mockSourceConn, util.NewTestLogger(), mockCache)
	client := &HubFeedClient{
		Feed: make(chan []byte),
	}

	//Act
	res := feedHub.Run()

	receivedMessage := ""

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for {
			msg := <-client.Feed
			receivedMessage = string(msg)
			wg.Done()
			return
		}
	}()

	feedHub.RegisterClient(client)
	go func() {
		mockSourceConn.RxMessages <- []byte("test")
	}()

	wg.Wait()

	//Assert
	assert.True(t, res)
	assert.True(t, mockSourceConn.ConnectCalled)
	assert.True(t, mockSourceConn.ListenCalled)
	assert.True(t, feedHub.IsRunning())

	assert.Equal(t, "test", receivedMessage)

	assert.True(t, mockCache.AddMessageCalled)
	assert.True(t, mockCache.GetMessagesCalled)
	assert.Equal(t, "test", string(mockCache.LastMessage))
	close(mockSourceConn.RxMessages)
}

func TestFeedHub_Stop_not_connected(t *testing.T) {
	//Arrange
	mockSourceConn := &mockSourceConnection{
		NextConnect: true,
	}
	feedHub := NewFeedHub(mockSourceConn, util.NewTestLogger(), nil)

	//Act
	feedHub.Stop()

	//Assert
	assert.True(t, mockSourceConn.GetReadyStateCalled)
	assert.False(t, mockSourceConn.DisconnectCalled)
}

func TestFeedHub_Stop_connected(t *testing.T) {
	//Arrange
	mockSourceConn := &mockSourceConnection{
		NextConnect:    true,
		NextReadyState: defs.ConnectionState.Open,
	}
	feedHub := NewFeedHub(mockSourceConn, util.NewTestLogger(), nil)

	//Act
	feedHub.Stop()

	//Assert
	assert.True(t, mockSourceConn.GetReadyStateCalled)
	assert.True(t, mockSourceConn.DisconnectCalled)
}

func TestFeedHub_Register_sends_cached_messages_to_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	mockCache := &message.MockCache{
		NextMessages: [][]byte{
			[]byte("message 1"),
			[]byte("message 2"),
		},
	}
	feedHub := &feedHubImpl{
		clients:      make(map[*HubFeedClient]bool),
		log:          util.NewTestLogger(),
		messageCache: mockCache,
	}

	client := &HubFeedClient{
		Feed: make(chan []byte),
	}

	receivedMessages := make([][]byte, 0)

	go func() {
		for {
			if message, ok := <-client.Feed; ok {
				receivedMessages = append(receivedMessages, message)
			} else {
				return
			}
		}
	}()

	//Act
	feedHub.RegisterClient(client)
	<-time.After(time.Second)

	//Assert
	assert.Equal(t, 2, len(receivedMessages))
	assert.Contains(t, receivedMessages, []byte("message 1"))
	assert.Contains(t, receivedMessages, []byte("message 2"))
	assert.True(t, mockCache.GetMessagesCalled)
	close(client.Feed)
}

func TestFeedHub_Register_Unregister_client(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)
	feedHub := &feedHubImpl{
		clients: make(map[*HubFeedClient]bool),
		log:     util.NewTestLogger(),
	}
	client := &HubFeedClient{
		Feed: make(chan []byte),
	}

	//Act//Assert
	feedHub.RegisterClient(client)
	assert.Equal(t, 1, len(feedHub.clients))

	feedHub.UnregisterClient(client)
	assert.Equal(t, 0, len(feedHub.clients))
}

func TestFeedHub_GetWebsocketStatus(t *testing.T) {
	//Arrange
	defer goleak.VerifyNone(t, ignoreOpenCensus)

	feedHub := &feedHubImpl{
		log:       util.NewTestLogger(),
		clients:   make(map[*HubFeedClient]bool),
		broadcast: make(chan []byte),
		source: &mockSourceConnection{
			NextReadyState: "open",
			NextFeedURL:    "test feed",
		},
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		for i := 0; i < 4; i++ {
			client := NewHubFeedClient(false)
			feedHub.RegisterClient(&client)
			wg.Done()
		}
	}()

	//Act//Assert
	res := feedHub.GetWebsocketStatus()
	assert.Equal(t, "open", res.ReadyState)
	assert.Equal(t, "test feed", res.RemoteFeedUrl)
	assert.Equal(t, uint32(0), res.ConnectedClients)
	wg.Wait()

	assert.Equal(t, uint32(4), feedHub.GetWebsocketStatus().ConnectedClients)
}
