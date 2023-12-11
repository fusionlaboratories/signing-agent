package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/keymanagement"
	"github.com/oracle/oci-go-sdk/v65/secrets"
	"github.com/oracle/oci-go-sdk/v65/vault"
	"github.com/pkg/errors"

	"github.com/qredo/signing-agent/internal/config"
	"github.com/qredo/signing-agent/internal/defs"
)

/*

	Setting up the credentials for OCI SDK -> https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdkconfig.htm
	Once you have the config and private key file, either put it in the default location of ~/.oci
	or set a env var OCI_CONFIG_FILE to point to the custom location of the config file.

	More info can be found at:
	  https://docs.oracle.com/en-us/iaas/Content/API/Concepts/sdk_authentication_methods.htm
      https://docs.oracle.com/en-us/iaas/api/#/en/secretmgmt/20180608/
      https://github.com/oracle/oci-go-sdk/blob/81ca54bf25c380aa231bbd166eb3ebdcbb825b50/example/example_databasetools_test.go#L702

*/

type OciStore struct {
	lock                     sync.RWMutex
	compartment_id           string
	vault_id                 string
	secret_encryption_key_id string
	config_secret            string
	vaults_client            vault.VaultsClient
	secrets_client           secrets.SecretsClient
	kms_crypto_client        keymanagement.KmsCryptoClient
}

func NewOciStore(cfg config.OciConfig) KVStore {
	s := &OciStore{
		compartment_id:           cfg.Compartment,
		vault_id:                 cfg.Vault,
		secret_encryption_key_id: cfg.SecretEncryptionKey,
		config_secret:            cfg.ConfigSecret,
		lock:                     sync.RWMutex{},
	}

	return s
}

func (s *OciStore) Get(key string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	secret_data, err := s.getSecret(s.config_secret)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secret_data) > 0 {
		err = json.Unmarshal(secret_data, &cfg)
		if err != nil {
			return nil, err
		}
	}

	if val, ok := cfg[key]; ok {
		return val, nil
	}

	return nil, defs.ErrKVNotFound
}

func (s *OciStore) Set(key string, data []byte) error {
	s.lock.RLock()

	secret_data, err := s.getSecret(s.config_secret)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secret_data) > 0 {
		err = json.Unmarshal(secret_data, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	cfg[key] = data

	secret_data, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := s.setSecret(s.config_secret, secret_data); err != nil {
		return err
	}

	return nil
}

func (s *OciStore) Del(key string) error {
	s.lock.RLock()

	secret_data, err := s.getSecret(s.config_secret)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secret_data) > 0 {
		err = json.Unmarshal(secret_data, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(cfg, key)

	secret_data, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := s.setSecret(s.config_secret, secret_data); err != nil {
		return err
	}

	return nil
}

func (s *OciStore) Init() error {
	vaults_client, err := vault.NewVaultsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		return err
	}
	s.vaults_client = vaults_client

	secrets_client, err := secrets.NewSecretsClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		return err
	}
	s.secrets_client = secrets_client

	kms_vault_client, err := keymanagement.NewKmsVaultClientWithConfigurationProvider(common.DefaultConfigProvider())
	if err != nil {
		return err
	}
	vault, err := kms_vault_client.GetVault(context.Background(), keymanagement.GetVaultRequest{
		VaultId: &s.vault_id,
	})
	if err != nil {
		return err
	}

	kms_crypto_client, err := keymanagement.NewKmsCryptoClientWithConfigurationProvider(common.DefaultConfigProvider(), *vault.CryptoEndpoint)
	if err != nil {
		return err
	}
	s.kms_crypto_client = kms_crypto_client

	return nil
}

func (s *OciStore) setSecret(name string, value []byte) error {
	var err error

	b64_encoded_secret := base64.StdEncoding.EncodeToString(value)
	if err != nil {
		return err
	}

	encrypted_secret, err := s.encryptSecret(b64_encoded_secret)
	if err != nil {
		return err
	}

	secret_id, err := s.getSecretId(name)
	if err != nil || secret_id == "" {
		_, err = s.vaults_client.CreateSecret(context.Background(), vault.CreateSecretRequest{
			CreateSecretDetails: vault.CreateSecretDetails{
				CompartmentId: &s.compartment_id,
				SecretContent: vault.Base64SecretContentDetails{
					Content: &encrypted_secret,
				},
				SecretName: &name,
				VaultId:    &s.vault_id,
				KeyId:      &s.secret_encryption_key_id,
			},
		})
	} else {
		_, err = s.vaults_client.UpdateSecret(context.Background(), vault.UpdateSecretRequest{
			SecretId: &secret_id,
			UpdateSecretDetails: vault.UpdateSecretDetails{
				SecretContent: vault.Base64SecretContentDetails{
					Content: &encrypted_secret,
				},
			},
		})
	}

	if err != nil {
		return err
	}

	s.waitForSecret(name)

	return nil
}

func (s *OciStore) getSecret(name string) ([]byte, error) {
	secret_id, err := s.getSecretId(name)
	if err != nil {
		return nil, nil
	}

	secret_response, err := s.secrets_client.GetSecretBundle(context.Background(),
		secrets.GetSecretBundleRequest{
			SecretId: &secret_id,
		},
	)
	if err != nil {
		return nil, err
	}

	secret_value := secret_response.SecretBundleContent.(secrets.Base64SecretBundleContentDetails).Content
	decrypted_secret, err := s.decryptSecret(*secret_value)
	if err != nil {
		return nil, nil
	}

	decoded_secret, err := base64.StdEncoding.DecodeString(decrypted_secret)
	if err != nil {
		return nil, nil
	}

	return decoded_secret, nil
}

func (s *OciStore) encryptSecret(value string) (string, error) {
	encrypt_response, err := s.kms_crypto_client.Encrypt(context.Background(), keymanagement.EncryptRequest{
		EncryptDataDetails: keymanagement.EncryptDataDetails{
			KeyId:     &s.secret_encryption_key_id,
			Plaintext: &value,
		},
	})

	if err != nil {
		return "", err
	}

	return *encrypt_response.Ciphertext, nil
}

func (s *OciStore) decryptSecret(value string) (string, error) {
	decrypt_response, err := s.kms_crypto_client.Decrypt(context.Background(), keymanagement.DecryptRequest{
		DecryptDataDetails: keymanagement.DecryptDataDetails{
			KeyId:      &s.secret_encryption_key_id,
			Ciphertext: &value,
		},
	})

	if err != nil {
		return "", err
	}

	return *decrypt_response.Plaintext, nil
}

func (s *OciStore) getSecretSummary(name string) (*vault.SecretSummary, error) {
	list_response, err := s.vaults_client.ListSecrets(context.Background(), vault.ListSecretsRequest{
		CompartmentId: &s.compartment_id,
		VaultId:       &s.vault_id,
		Name:          &name,
	})
	if err != nil {
		return nil, err
	}

	if len(list_response.Items) < 1 {
		return nil, errors.New("not found")
	}

	return &list_response.Items[0], nil
}

func (s *OciStore) getSecretId(name string) (string, error) {
	summary, err := s.getSecretSummary(name)
	if err != nil {
		return "", err
	}

	return *summary.Id, nil
}

func (s *OciStore) waitForSecret(name string) {
	for {
		summary, err := s.getSecretSummary(name)
		if err != nil {
			continue
		}

		if summary.LifecycleState == vault.SecretSummaryLifecycleStateActive {
			return
		}
	}
}
