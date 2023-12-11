package action

type MockSigner struct {
	ActionApproveCalled        bool
	ActionRejectCalled         bool
	ApproveActionMessageCalled bool
	SetKeyCalled               bool

	LastBlsPrivateKey string
	LastActionId      string
	LastMessage       []byte
	NextError         error
	NextSetKeyError   error

	Counter int
}

func (m *MockSigner) SetKey(blsPrivateKey string) error {
	m.SetKeyCalled = true
	m.LastBlsPrivateKey = blsPrivateKey
	return m.NextSetKeyError
}
func (m *MockSigner) ActionApprove(actionID string) error {
	m.ActionApproveCalled = true
	m.LastActionId = actionID
	return m.NextError
}
func (m *MockSigner) ActionReject(actionID string) error {
	m.ActionRejectCalled = true
	m.LastActionId = actionID
	return m.NextError
}
func (m *MockSigner) ApproveActionMessage(actionID string, message []byte) error {
	m.ApproveActionMessageCalled = true
	m.LastActionId = actionID
	m.LastMessage = message
	m.Counter++
	return m.NextError
}
