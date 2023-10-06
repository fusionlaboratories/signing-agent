package e2e

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

func createAWSConfig() *config.Config {
	pwd, _ := os.Getwd()
	f, err := os.ReadFile(pwd + "/../../testdata/e2e/aws_config.yaml")
	if err != nil {
		log.Fatalf("error reading test config file: %v", err)
	}

	var testCfg struct {
		Region     string `yaml:"region"`
		SecretName string `yaml:"configSecret"`
	}
	err = yaml.Unmarshal(f, &testCfg)
	if err != nil {
		log.Fatalf("error unmarshaling test config file: %v", err)
	}

	cfg := &config.Config{
		Store: config.Store{
			Type: "aws",
			AwsConfig: config.AWSConfig{
				Region:     testCfg.Region,
				SecretName: testCfg.SecretName,
			},
		},
	}

	return cfg
}

// TestAWSStoreInitDoesNotReturnError confirms initialising the store is OK (error == nil).
func TestAWSStoreInitDoesNotReturnError(t *testing.T) {
	cfg := createAWSConfig()
	store := util.CreateStore(cfg)

	err := store.Init()
	assert.Nil(t, err)
}

// TestAWSStoreGetReturnsNotFound confirms getting an non-existent key returns a "not found" error.
func TestAWSStoreGetReturnsNotFound(t *testing.T) {
	cfg := createAWSConfig()
	store := util.CreateStore(cfg)
	err := store.Init()
	assert.Nil(t, err)

	_, err = store.Get("some_unknown_key")
	assert.Equal(t, "not found", err.Error())
}

// TestAWSStoreGetReturnsSetValue checks setting and getting a new key/value by adding a k/v to the store, reading it
// back, and confirming the value is the same.
func TestAWSStoreGetReturnsSetValue(t *testing.T) {
	cfg := createAWSConfig()
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

// TestAWSStoreDeleteKey checks deleting a key/value.
func TestAWSStoreDeleteKey(t *testing.T) {
	// Arrange
	cfg := createAWSConfig()
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
