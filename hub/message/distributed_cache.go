package message

import (
	"context"
	"encoding/json"

	"github.com/qredo/signing-agent/lib/store"
	"go.uber.org/zap"
)

const (
	keyPrefix  = "transaction:"
	keyPattern = "transaction:*"
)

var ctx = context.Background()

// distributedCache is a messages cache to be used in multi-instance Signing Agent
type distributedCache struct {
	kvStore store.KVStore
	log     *zap.SugaredLogger
}

// AddMessage stores a message into the cache
func (c *distributedCache) AddMessage(message []byte) {
	info := messageInfo{}
	if err := json.Unmarshal(message, &info); err != nil {
		c.log.Debugf("message Cache: error [%v] while unmarshaling the message [%s]", err, string(message))
	} else {
		expiration := info.getExpiration()
		//add only not expired messages
		if expiration > 0 {
			c.kvStore.Set(ctx, c.getKey(info.ID), string(message), expiration)
		}
	}
}

// GetMessages returns all received and not expired messages. All expired or invalid messages will be cleared out
func (c *distributedCache) GetMessages() [][]byte {
	messages := make([][]byte, 0)

	iter := c.kvStore.Scan(ctx, 0, keyPattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		msg, err := c.kvStore.Get(ctx, key).Result()
		if err != nil {
			c.log.Errorf("message Cache: error while retrieving messages: %v", err)
		} else {
			messages = append(messages, []byte(msg))
		}
	}

	if err := iter.Err(); err != nil {
		c.log.Errorf("message Cache: error while retrieving messages: %v", err)
	}

	return messages
}

// RemoveMessage deletes the message for the given ID
func (c *distributedCache) RemoveMessage(ID string) {
	c.log.Debugf("message Cache: removing message with ID `%s`", ID)
	c.kvStore.Del(ctx, c.getKey(ID))
}

func (c *distributedCache) getKey(ID string) string {
	return keyPrefix + ID
}
