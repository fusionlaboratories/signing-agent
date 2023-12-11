package store

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qredo/signing-agent/internal/util"
)

var TestDataDBStoreFilePath = "../../testdata/test-store.db"

func TestStorage(t *testing.T) {
	kv := util.NewFileStore(TestDataDBStoreFilePath)
	err := kv.Init()
	defer func() {
		err = os.Remove(TestDataDBStoreFilePath)
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)

	store := NewAgentStore(kv)

	t.Run(
		"get agent info - agent not set",
		func(t *testing.T) {
			agentInfo, err := store.GetAgentInfo()
			assert.Nil(t, agentInfo)
			assert.Nil(t, err)
		})

	t.Run(
		"register agent - invalid id",
		func(t *testing.T) {
			agentID := ""
			err = store.SaveAgentInfo(agentID, &AgentInfo{})

			assert.NotNil(t, err)
			assert.Equal(t, "invalid agentID", err.Error())
		})

	t.Run(
		"register agent - invalid agent info",
		func(t *testing.T) {
			agentID := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
			err = store.SaveAgentInfo(agentID, nil)

			assert.NotNil(t, err)
			assert.Equal(t, "invalid agent info", err.Error())
		})

	t.Run(
		"register agent - saves agent info",
		func(t *testing.T) {
			agentID := "5zPWqLZaPqAaNenjyzWy5rcaGm4PuT1bfP74GgrzFUJn"
			newAgentInfo := &AgentInfo{
				BLSPrivateKey: "bls priv key",
				ECPrivateKey:  "ec priv key",
				WorkspaceID:   "wkpID",
				APIKeyID:      "api key id",
				APIKeySecret:  "some secret",
			}
			err = store.SaveAgentInfo(agentID, newAgentInfo)

			assert.Nil(t, err)

			agentInfo, err := store.GetAgentInfo()
			assert.Nil(t, err)
			assert.Equal(t, *agentInfo, *newAgentInfo)
		})
}
