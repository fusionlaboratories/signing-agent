package message

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
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

	data := messageInfo{
		ID:         "test",
		ExpireTime: time.Now().Unix() + 60,
	}
	testData, _ := json.Marshal(data)
	//Act
	sut.AddMessage(testData)

	//Assert
	assert.Equal(t, 1, len(sut.messages))
	assert.Contains(t, sut.messages, "test")
	assert.Equal(t, testData, sut.messages["test"])
}

func TestLocalCache_GetMessages_returns_valid_clears_expired_and_invalid(t *testing.T) {
	//Arrange
	actionExpired := &defs.ActionInfo{
		ID:         "expired",
		Status:     1,
		ExpireTime: 13285485,
	}
	validAction := &defs.ActionInfo{
		ID:         "valid",
		Status:     1,
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
	assert.Contains(t, string(res[0]), "\"id\":\"valid\",\"status\":1,\"messages\":null,\"expireTime\"")
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
