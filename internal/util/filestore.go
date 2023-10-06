package util

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/qredo/signing-agent/internal/defs"
)

func NewFileStore(fileName string) KVStore {
	fs := &FileStore{
		fileName: fileName,
		data:     map[string][]byte{},
	}

	return fs
}

type FileStore struct {
	sync.RWMutex
	fileName string
	data     map[string][]byte
}

func (s *FileStore) Init() error {
	s.Lock()
	defer s.Unlock()

	b, err := os.ReadFile(s.fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return s.save()
		}
		return err
	}

	if err := json.Unmarshal(b, &s.data); err != nil {
		return err
	}

	return nil
}

// caller must handle concurrency
func (s *FileStore) save() error {

	b, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return os.WriteFile(s.fileName, b, 0600)
}

func (s *FileStore) Get(key string) ([]byte, error) {
	s.RLock()
	defer s.RUnlock()

	if data, ok := s.data[key]; ok {
		return data, nil
	}

	return nil, defs.ErrKVNotFound
}

func (s *FileStore) Set(key string, data []byte) error {
	s.Lock()
	defer s.Unlock()

	s.data[key] = data

	return s.save()
}

func (s *FileStore) Del(key string) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.data[key]; ok {
		delete(s.data, key)
		return s.save()
	}

	return nil
}
