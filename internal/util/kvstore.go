package util

// KVStore is an interface to a simple key-value store used by the lib
type KVStore interface {
	// Get returns the data for given key. If key is not found, return nil, defs.ErrKVNotFound
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Del(key string) error
	Init() error
}
