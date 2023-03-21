package hub

import (
	"net/http"
	"sync"
	"time"
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

type MockWebsocketUpgrader struct {
	UpgradeCalled           bool
	NextError               error
	NextWebsocketConnection WebsocketConnection
	LastWriter              http.ResponseWriter
	LastRequest             *http.Request
	LastResponseHeader      http.Header
}

func (m *MockWebsocketUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketConnection, error) {
	m.UpgradeCalled = true
	m.LastWriter = w
	m.LastRequest = r
	m.LastResponseHeader = responseHeader
	return m.NextWebsocketConnection, m.NextError
}

type MockWebsocketConnection struct {
	ReadMessageCalled      bool
	WriteMessageCalled     bool
	CloseCalled            bool
	WriteControlCalled     bool
	SetReadDeadlineCalled  bool
	SetPongHandlerCalled   bool
	SetPingHandlerCalled   bool
	SetWriteDeadlineCalled bool
	LastMessageType        int
	NextMessageType        int
	LastData               []byte
	NextData               []byte
	NextError              error
	read                   chan bool
}

func (m *MockWebsocketConnection) ReadMessage() (messageType int, p []byte, err error) {
	m.ReadMessageCalled = true
	<-m.read

	return m.NextMessageType, m.NextData, m.NextError
}

func (m *MockWebsocketConnection) WriteMessage(messageType int, data []byte) error {
	m.WriteMessageCalled = true
	m.LastMessageType = messageType
	m.LastData = data
	return m.NextError
}

func (m *MockWebsocketConnection) Close() error {
	m.CloseCalled = true
	return m.NextError
}

func (m *MockWebsocketConnection) WriteControl(messageType int, data []byte, deadline time.Time) error {
	m.WriteControlCalled = true
	m.LastMessageType = messageType

	return m.NextError
}
func (m *MockWebsocketConnection) SetReadDeadline(t time.Time) error {
	m.SetReadDeadlineCalled = true

	return nil
}

func (m *MockWebsocketConnection) SetPongHandler(h func(appData string) error) {
	m.SetPongHandlerCalled = true
}

func (m *MockWebsocketConnection) SetPingHandler(h func(appData string) error) {
	m.SetPingHandlerCalled = true
}

func (m *MockWebsocketConnection) SetWriteDeadline(t time.Time) error {
	m.SetWriteDeadlineCalled = true

	return m.NextError
}

type mockSourceConnection struct {
	ConnectCalled       bool
	ListenCalled        bool
	DisconnectCalled    bool
	GetReadyStateCalled bool
	NextConnect         bool
	NextReadyState      string
	RxMessages          chan []byte
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
	return ""
}

func (m *mockSourceConnection) GetReadyState() string {
	m.GetReadyStateCalled = true
	return m.NextReadyState
}

func (m *mockSourceConnection) GetSendChannel() chan []byte {
	return m.RxMessages
}
