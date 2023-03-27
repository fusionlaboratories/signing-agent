package message

import (
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type messageInfo struct {
	ID         string `json:"id"`
	ExpireTime int64  `json:"expireTime"`
}

func (m *messageInfo) getExpiration() time.Duration {
	return time.Duration(m.ExpireTime-time.Now().Unix()) * time.Second
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

func NewCacher(isMultiInstance bool, log *zap.SugaredLogger, kvStore *redis.Client) Cacher {
	if isMultiInstance {
		return &distributedCache{
			kvStore: kvStore,
			log:     log,
		}
	}

	return &localCache{
		log:      log,
		lock:     sync.RWMutex{},
		messages: make(map[string][]byte),
	}
}
