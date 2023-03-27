package store

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type KVStore interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type KVStoreMock struct {
	GetCalled  bool
	SetCalled  bool
	ScanCalled bool
	DelCalled  bool

	NextStringCmd *redis.StringCmd
	NextStatusCmd *redis.StatusCmd
	NextScanCmd   *redis.ScanCmd

	LastKey        string
	LastValue      interface{}
	LastExpiration time.Duration
	LastCursor     uint64
	LastMatch      string
	LastCount      int64
}

func (m *KVStoreMock) Get(ctx context.Context, key string) *redis.StringCmd {
	m.GetCalled = true
	m.LastKey = key
	return m.NextStringCmd
}
func (m *KVStoreMock) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.SetCalled = true
	m.LastKey = key
	m.LastValue = value
	m.LastExpiration = expiration
	return m.NextStatusCmd
}

func (m *KVStoreMock) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	m.ScanCalled = true
	m.LastCursor = cursor
	m.LastMatch = match
	m.LastCount = count
	return m.NextScanCmd
}

func (m *KVStoreMock) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	m.DelCalled = true
	m.LastKey = keys[0]
	return nil
}
