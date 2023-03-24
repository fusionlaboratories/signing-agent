package message

import (
	"encoding/json"
	"sync"

	"github.com/qredo/signing-agent/defs"
	"go.uber.org/zap"
)

// localCache is a messages cache to be used in single instance Signing Agent
type localCache struct {
	messages map[string][]byte
	lock     sync.RWMutex
	log      *zap.SugaredLogger
}

// AddMessage stores a message into the cache
func (c *localCache) AddMessage(message []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	data := dataID{}
	if err := json.Unmarshal(message, &data); err != nil {
		c.log.Debugf("message Cache: error [%v] while unmarshaling the message [%s]", err, string(message))
	} else {
		c.messages[data.ID] = message
	}
}

// GetMessages returns all received and not expired messages. All expired or invalid messages will be cleared out
func (c *localCache) GetMessages() [][]byte {
	c.lock.Lock()
	defer c.lock.Unlock()

	res := make([][]byte, 0)
	for id, message := range c.messages {
		action := &defs.ActionInfo{}

		err := json.Unmarshal(message, &action)
		if err != nil || action.IsExpired() {
			delete(c.messages, id)
		} else {
			res = append(res, message)
		}
	}

	return res
}

// RemoveMessage deletes the message for the given ID
func (c *localCache) RemoveMessage(ID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.log.Debugf("message Cache: removing message with ID `%s`", ID)
	delete(c.messages, ID)
}
