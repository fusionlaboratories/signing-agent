package e2e

//TODO - to be fixed within the task for all integration + e2e tests
/*
import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/util"
)

func createGCPConfig() config.Config {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(pwd + "/../../testdata/integration/gcp_config.yaml")
	if err != nil {
		log.Fatalf("error reading test config file: %v", err)
	}

	var testCfg struct {
		ProjectID  string `yaml:"projectID"`
		SecretName string `yaml:"configSecret"`
	}
	err = yaml.Unmarshal(f, &testCfg)
	if err != nil {
		log.Fatalf("error unmarshaling test config file: %v", err)
	}

	cfg := config.Config{
		Store: config.Store{
			Type: "gcp",
			GcpConfig: config.GCPConfig{
				ProjectID:  testCfg.ProjectID,
				SecretName: testCfg.SecretName,
			},
		},
	}

	return cfg
}

// TestGCPStoreInitDoesNotReturnError tests initialisation of GCP store.
func TestGCPStoreInitDoesNotReturnError(t *testing.T) {
	cfg := createGCPConfig()
	store := util.CreateStore(cfg)

	err := store.Init()
	assert.Nil(t, err)
}

// TestGCPStoreGetReturnsNotFound checks that a not_found error is returned when looking up of non-existent key.
func TestGCPStoreGetReturnsNotFound(t *testing.T) {
	cfg := createGCPConfig()
	store := util.CreateStore(cfg)

	err := store.Init()
	assert.Nil(t, err)

	_, err = store.Get("some_unknown_key")
	assert.Equal(t, "not found", err.Error())
}

// TestGCPStoreGetReturnsSetValue checks setting and getting a new key/value by adding a k/v to the store, reading it
// back, and confirming the value is the same.
func TestGCPStoreGetReturnsSetValue(t *testing.T) {
	cfg := createGCPConfig()
	store := util.CreateStore(cfg)
	err := store.Init()
	assert.Nil(t, err)

	key := "PoPCorn"
	value := "sweet or salty?"
	err = store.Set(key, []byte(value))
	assert.Nil(t, err)

	<-time.After(time.Second)

	result, err := store.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, string(result))
}

// TestGCPStoreDeleteKey checks deleting a key/value.
func TestGCPStoreDeleteKey(t *testing.T) {
	cfg := createGCPConfig()
	store := util.CreateStore(cfg)
	err := store.Init()
	assert.Nil(t, err)

	// add a new key/value
	key := "PoPCorn"
	value := "sweet or salty?"
	err = store.Set(key, []byte(value))
	assert.Nil(t, err)

	<-time.After(time.Second)

	// check the key exists
	result, err := store.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, value, string(result))

	// delete a key and confirm it no-longer exists
	err = store.Del(key)
	assert.Nil(t, err)
	_, err = store.Get(key)
	assert.Equal(t, "not found", err.Error())
}
*/
