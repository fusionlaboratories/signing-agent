package message

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/util"
	"github.com/stretchr/testify/assert"
)

func TestLocalCache_AddMessage_fails_to_unmarshal(t *testing.T) {
	//Arrange
	sut := &localCache{
		log:      util.NewTestLogger(),
		lock:     sync.RWMutex{},
		messages: make(map[string][]byte),
	}

	//Act
	sut.AddMessage([]byte("some message"))

	//Assert
	assert.Empty(t, sut.messages)
}

func TestLocalCache_AddMessage_adds_message(t *testing.T) {
	//Arrange
	sut := &localCache{
		log:      util.NewTestLogger(),
		lock:     sync.RWMutex{},
		messages: make(map[string][]byte),
	}

	//Act
	sut.AddMessage([]byte(`{"id":"some id", "data":"test"}`))

	//Assert
	assert.Equal(t, 1, len(sut.messages))
	assert.Contains(t, sut.messages, "some id")
	assert.Equal(t, `{"id":"some id", "data":"test"}`, string(sut.messages["some id"]))
}

func TestLocalCache_GetMessages_returns_valid_clears_expired_and_invalid(t *testing.T) {
	//Arrange
	actionExpired := &defs.ActionInfo{
		ID:         "expired",
		AgentID:    "agent id",
		Type:       "type",
		Status:     "pending",
		Timestamp:  121475,
		ExpireTime: 13285485,
	}
	validAction := &defs.ActionInfo{
		ID:         "valid",
		AgentID:    "agent id",
		Type:       "type",
		Status:     "pending",
		Timestamp:  121475,
		ExpireTime: time.Now().Unix() + 10,
	}

	expiredActionJson, _ := json.Marshal(actionExpired)
	validActionJson, _ := json.Marshal(validAction)
	sut := &localCache{
		log:  util.NewTestLogger(),
		lock: sync.RWMutex{},
		messages: map[string][]byte{
			"invalid": []byte("{bla bla}"),
			"expired": expiredActionJson,
			"valid":   validActionJson,
		},
	}

	//Act
	res := sut.GetMessages()

	//Assert
	assert.Equal(t, 1, len(sut.messages))
	assert.Contains(t, sut.messages, "valid")
	assert.Equal(t, 1, len(res))
	assert.Contains(t, string(res[0]), "{\"id\":\"valid\",\"coreClientID\":\"agent id\",\"type\":\"type\",\"status\":\"pending\",\"timestamp\":121475")
}

func TestLocalCache_RemoveMessage(t *testing.T) {
	//Arrange
	sut := &localCache{
		log:  util.NewTestLogger(),
		lock: sync.RWMutex{},
		messages: map[string][]byte{
			"test": []byte(""),
		},
	}

	//Act
	sut.RemoveMessage("test")

	//Assert
	assert.Empty(t, sut.messages)
}
