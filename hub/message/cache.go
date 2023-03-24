package message

import (
	"sync"

	"go.uber.org/zap"
)

type dataID struct {
	ID string `json:"id"`
}

type CacheRemover interface {
	RemoveMessage(ID string)
}

type Cache interface {
	AddMessage(message []byte)
	GetMessages() [][]byte
}

type Cacher interface {
	CacheRemover
	Cache
}

func NewCacher(isMultiInstance bool, log *zap.SugaredLogger) Cacher {
	if isMultiInstance {
		//TODO
		return nil
	}

	return &localCache{
		log:      log,
		lock:     sync.RWMutex{},
		messages: make(map[string][]byte),
	}
}
