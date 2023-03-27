package message

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/qredo/signing-agent/lib/store"
	"github.com/qredo/signing-agent/util"
	"github.com/stretchr/testify/assert"
)

func TestDistributedCache_AddMessage_fails_to_unmarshal_doesnt_add(t *testing.T) {
	//Arrange
	mockKVStore := &store.KVStoreMock{}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	//Act
	sut.AddMessage([]byte("invalid"))

	//Assert
	assert.False(t, mockKVStore.SetCalled)
}

func TestDistributedCache_AddMessage_expired_doesnt_add(t *testing.T) {
	//Arrange
	mockKVStore := &store.KVStoreMock{}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	message, _ := json.Marshal(messageInfo{
		ID:         "someID",
		ExpireTime: 1234,
	})

	//Act
	sut.AddMessage(message)

	//Assert
	assert.False(t, mockKVStore.SetCalled)
}

func TestDistributedCache_AddMessage_adds_valid_message(t *testing.T) {
	//Arrange
	mockKVStore := &store.KVStoreMock{}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	message, _ := json.Marshal(messageInfo{
		ID:         "someID",
		ExpireTime: time.Now().Unix() + 100,
	})

	//Act
	sut.AddMessage(message)

	//Assert
	assert.True(t, mockKVStore.SetCalled)
	assert.Equal(t, "transaction:someID", mockKVStore.LastKey)
	assert.Contains(t, mockKVStore.LastValue, "{\"id\":\"someID\",\"expireTime\":")
	assert.NotEmpty(t, mockKVStore.LastExpiration)
}

func TestDistributedCache_GetMessages_scan_error(t *testing.T) {
	//Arrange
	scanCmd := &redis.ScanCmd{}
	scanCmd.SetErr(errors.New("some scan error"))

	mockKVStore := &store.KVStoreMock{
		NextScanCmd: scanCmd,
	}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	//Act
	res := sut.GetMessages()

	//Assert
	assert.Empty(t, res)
	assert.True(t, mockKVStore.ScanCalled)
	assert.Empty(t, mockKVStore.LastCursor)
	assert.Equal(t, "transaction:*", mockKVStore.LastMatch)
	assert.Empty(t, mockKVStore.LastCount)
}

func TestDistributedCache_GetMessages_get_error(t *testing.T) {
	//Arrange
	scanCmd := &redis.ScanCmd{}
	scanCmd.SetVal([]string{"test1", "test2"}, 0)

	getCmd := &redis.StringCmd{}
	getCmd.SetErr(errors.New("some error"))

	mockKVStore := &store.KVStoreMock{
		NextScanCmd:   scanCmd,
		NextStringCmd: getCmd,
	}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	//Act
	res := sut.GetMessages()

	//Assert
	assert.Empty(t, res)
	assert.True(t, mockKVStore.ScanCalled)
	assert.Empty(t, mockKVStore.LastCursor)
	assert.Equal(t, "transaction:*", mockKVStore.LastMatch)
	assert.Empty(t, mockKVStore.LastCount)
}

func TestDistributedCache_GetMessages_retrieves_messages(t *testing.T) {
	//Arrange
	scanCmd := &redis.ScanCmd{}
	scanCmd.SetVal([]string{"test1", "test2"}, 0)

	getCmd := &redis.StringCmd{}
	getCmd.SetVal("messages")

	mockKVStore := &store.KVStoreMock{
		NextScanCmd:   scanCmd,
		NextStringCmd: getCmd,
	}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	//Act
	res := sut.GetMessages()

	//Assert
	assert.NotEmpty(t, res)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, "messages", string(res[0]))
	assert.Equal(t, "messages", string(res[1]))
	assert.True(t, mockKVStore.ScanCalled)
	assert.Empty(t, mockKVStore.LastCursor)
	assert.Equal(t, "transaction:*", mockKVStore.LastMatch)
	assert.Empty(t, mockKVStore.LastCount)
}

func TestDistributedCache_RemoveMessage(t *testing.T) {
	//Arrange
	mockKVStore := &store.KVStoreMock{}
	sut := distributedCache{
		log:     util.NewTestLogger(),
		kvStore: mockKVStore,
	}

	//Act
	sut.RemoveMessage("testID")

	//Assert
	assert.True(t, mockKVStore.DelCalled)
	assert.Equal(t, "transaction:testID", mockKVStore.LastKey)
}
