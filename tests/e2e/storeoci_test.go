package e2e

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/vault"
	"github.com/test-go/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/util"
)

func createOciConfig() *config.Config {
	pwd, _ := os.Getwd()
	os.Setenv("OCI_CONFIG_FILE", pwd+"/../../testdata/e2e/oci_config")

	compartment := os.Getenv("OCI_COMPARTMENT")
	vault := os.Getenv("OCI_VAULT")
	encryptionkey := os.Getenv("OCI_ENC_KEY")

	cfg := &config.Config{
		Store: config.Store{
			Type: "oci",
			OciConfig: config.OciConfig{
				Compartment:         compartment,
				Vault:               vault,
				SecretEncryptionKey: encryptionkey,
				ConfigSecret:        "e2e-" + strconv.FormatInt(time.Now().Unix(), 10),
			},
		},
	}

	return cfg
}

func deleteSecret(t *testing.T, cfg *config.Config) {
	vaults_client, err := vault.NewVaultsClientWithConfigurationProvider(common.DefaultConfigProvider())
	require.Nil(t, err)

	list_response, err := vaults_client.ListSecrets(context.Background(), vault.ListSecretsRequest{
		CompartmentId: &cfg.Store.OciConfig.Compartment,
		VaultId:       &cfg.Store.OciConfig.Vault,
		Name:          &cfg.Store.OciConfig.ConfigSecret,
	})
	require.Nil(t, err)

	secret_id := &list_response.Items[0].Id

	_, err = vaults_client.ScheduleSecretDeletion(context.Background(), vault.ScheduleSecretDeletionRequest{
		SecretId: *secret_id,
		ScheduleSecretDeletionDetails: vault.ScheduleSecretDeletionDetails{
			TimeOfDeletion: &common.SDKTime{
				Time: time.Now().Add(time.Hour * 25),
			},
		},
	})
	require.Nil(t, err)
}

func Test_E2E_OciStore_Init_does_not_return_an_error(t *testing.T) {
	// Arrange
	cfg := createOciConfig()
	sut := util.CreateStore(cfg)

	// Act
	err := sut.Init()

	// Assert
	assert.Nil(t, err)
}

func Test_E2E_OciStore_Get_returns_not_found(t *testing.T) {
	// Arrange
	cfg := createOciConfig()
	sut := util.CreateStore(cfg)

	err := sut.Init()
	require.Nil(t, err)

	// Act
	_, err = sut.Get("some_unknown_key")

	// Assert
	assert.Equal(t, "not found", err.Error())
}

func Test_E2E_OciStore_Get_returns_set_value(t *testing.T) {
	// Arrange
	cfg := createOciConfig()
	sut := util.CreateStore(cfg)

	err := sut.Init()
	require.Nil(t, err)

	value := "some value"

	// Act
	err = sut.Set("some_key", []byte(value))
	require.Nil(t, err)

	<-time.After(time.Second)

	result, err := sut.Get("some_key")

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, value, string(result))

	// Cleanup
	deleteSecret(t, cfg)
}

func Test_E2E_OciStore_Del_removes_specified_key_and_vallue(t *testing.T) {
	// Arrange
	cfg := createOciConfig()
	sut := util.CreateStore(cfg)
	value := "some value"

	err := sut.Init()
	if err != nil {
		goto cleanup
	}

	err = sut.Set("some_other_key", []byte(value))
	if err != nil {
		goto cleanup
	}

	<-time.After(time.Second)

	_, err = sut.Get("some_other_key")
	if err != nil {
		goto cleanup
	}

	// Act
	err = sut.Del("some_other_key")
	assert.Nil(t, err)

	<-time.After(time.Second)

	_, err = sut.Get("some_other_key")

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "not found", err.Error())

cleanup:
	// Cleanup
	deleteSecret(t, cfg)
}
