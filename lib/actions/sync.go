package actions

import (
	"github.com/go-redsync/redsync/v4"
)

type syncI interface {
	NewMutex(name string, options ...redsync.Option) *redsync.Mutex
}

type mockSync struct {
	NewMutexCalled bool
	LastName       string
	NextMutex      *redsync.Mutex
}

func (m *mockSync) NewMutex(name string, options ...redsync.Option) *redsync.Mutex {
	m.NewMutexCalled = true
	m.LastName = name
	return m.NextMutex
}

type mutex interface {
	Lock() error
	Unlock() (bool, error)
}
