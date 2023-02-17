package util

// KVStore is an interface to a simple key-value store used by the core lib
type KVStore interface {
	// Get returns the data for given key. If key is not found, return nil, defs.KVErrNotFound
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Del(key string) error
	Init() error
}
