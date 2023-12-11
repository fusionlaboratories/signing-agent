package message

type MockCache struct {
	AddMessageCalled    bool
	GetMessagesCalled   bool
	RemoveMessageCalled bool

	LastMessage []byte
	LastID      string

	NextMessages [][]byte
}

func (m *MockCache) AddMessage(message []byte) {
	m.AddMessageCalled = true
	m.LastMessage = message

}
func (m *MockCache) GetMessages() [][]byte {
	m.GetMessagesCalled = true
	return m.NextMessages
}

func (m *MockCache) RemoveMessage(ID string) {
	m.RemoveMessageCalled = true
	m.LastID = ID
}
