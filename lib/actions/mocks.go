package actions

type MockActionSyncronizer struct {
	ShouldHandleActionCalled bool
	AcquireLockCalled        bool
	ReleaseCalled            bool
	LastActionId             string
	NextShouldHandle         bool
	NextLockError            error
	NextReleaseError         error
}

func (m *MockActionSyncronizer) ShouldHandleAction(actionID string) bool {
	m.ShouldHandleActionCalled = true
	m.LastActionId = actionID
	return m.NextShouldHandle
}
func (m *MockActionSyncronizer) AcquireLock() error {
	m.AcquireLockCalled = true
	return m.NextLockError
}
func (m *MockActionSyncronizer) Release(actionID string) error {
	m.ReleaseCalled = true
	m.LastActionId = actionID
	return m.NextReleaseError
}

type mutexMock struct {
	LockCalled   bool
	UnlockCalled bool
	NextError    error
	NextUnlock   bool
}

func (m *mutexMock) Lock() error {
	m.LockCalled = true
	return m.NextError
}
func (m *mutexMock) Unlock() (bool, error) {
	m.UnlockCalled = true
	return m.NextUnlock, m.NextError
}
