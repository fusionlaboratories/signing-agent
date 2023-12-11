package hub

import (
	"time"
)

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
