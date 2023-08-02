package util

/*
In order to use GCP to store signing-agent BLS data, your GCP account should be configured to use the
GCP secret manager and a key created and initialised with the string "initialise me" (without the "s).

Note that both customer manage (CMEK) and google managed encryption keys can be used.

The store stanza in the signing-agent's configuration file should be set as follows:

  store:
    type: gcp
    gcp:
      projectID: [project ID where the configured secret resides]
      configSecret: [the name of the secret]

Google Cloud authentication must be established before starting the signing-agent. The authenticated credentials
should reflect the GCP location which contains the projectID and secret manager secret. The signing-agent
picks up and uses credentials from its local environment as it starts up.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/defs"
)

// GCPSecretNotInitialised is used to identify an uninitialised AWS secret.
const (
	GCPSecretNotInitialised string = "initialise me"
)

type GCPStore struct {
	lock       sync.RWMutex
	projectID  string
	locationID string
	secretName string
	svc        *secretmanager.Client
}

// NewGCPStore creates and returns the GCP KVStore.
func NewGCPStore(cfg config.GCPConfig) KVStore {
	s := &GCPStore{
		projectID:  cfg.ProjectID,
		secretName: cfg.SecretName,
		lock:       sync.RWMutex{},
	}

	return s
}

// Get returns the value of the named key, or error if not found.
func (s *GCPStore) Get(key string) ([]byte, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ctx := context.Background()
	secretData, err := s.getSecret(ctx)
	if err != nil {
		return nil, err
	}

	cfg := make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			return nil, err
		}
	}

	if val, ok := cfg[key]; ok {
		return val, nil
	}

	return nil, defs.KVErrNotFound
}

// Set adds/updates the named key with value in data.
func (s *GCPStore) Set(key string, data []byte) error {
	s.lock.RLock()

	ctx := context.Background()
	secretData, err := s.getSecret(ctx)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	cfg := make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	cfg[key] = data

	secretData, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := s.setSecret(ctx, secretData); err != nil {
		return err
	}

	return nil
}

// Del deletes the named key.
func (s *GCPStore) Del(key string) error {
	s.lock.RLock()

	ctx := context.Background()
	secretData, err := s.getSecret(ctx)
	if err != nil {
		s.lock.RUnlock()
		return err
	}

	var cfg map[string][]byte = make(map[string][]byte)
	if len(secretData) > 0 {
		err = json.Unmarshal(secretData, &cfg)
		if err != nil {
			s.lock.RUnlock()
			return err
		}
	}

	s.lock.RUnlock()
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(cfg, key)

	secretData, err = json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := s.setSecret(ctx, secretData); err != nil {
		return err
	}

	return nil
}

// Init sets up the GCP session and checks the connection by reading the secret.
func (s *GCPStore) Init() error {

	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot initialise GCP store")
	}

	s.svc = client

	err = s.initConnection(ctx)

	return err
}

// getSecret reads the secret with name from GCP.
func (s *GCPStore) getSecret(ctx context.Context) ([]byte, error) {
	result, err := s.readSecret(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lookup the secret")
	}

	return result.Payload.Data, nil
}

// readSecret reads and returns the latest version of the secret from GCP.
func (s *GCPStore) readSecret(ctx context.Context) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	versionReq := &secretmanagerpb.GetSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", s.projectID, s.secretName),
	}

	version, err := s.svc.GetSecretVersion(ctx, versionReq)
	if err != nil {
		return nil, err
	}

	// Build the request.
	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: version.Name,
	}

	// Call the API.
	result, err := s.svc.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return nil, errors.Wrap(err, "failed to access secret version")
	}
	return result, nil
}

// setSecret stores data in the named secret.
func (s *GCPStore) setSecret(ctx context.Context, data []byte) error {
	addSecretVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: fmt.Sprintf("projects/%s/secrets/%s", s.projectID, s.secretName),
		Payload: &secretmanagerpb.SecretPayload{
			Data: data,
		},
	}
	_, err := s.svc.AddSecretVersion(ctx, addSecretVersionReq)
	if err != nil {
		return err
	}
	return nil
}

// initConnection checks the GCP connection and that the named secret can be read. If the secret value is the
// string SecretNotInitialised, the secret is initialised to SecretInitialised.
func (s *GCPStore) initConnection(ctx context.Context) error {
	// check named secret exists
	result, err := s.readSecret(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot initialise GCP connection")
	}

	// payload should either be uninitialised or, already initialised, JSON data.
	if result.Payload.Data != nil {

		// check if JSON and return
		var js interface{}
		err := json.Unmarshal([]byte(result.Payload.Data), &js)
		if err == nil {
			return nil
		}

		// check if uninitialised and initialise it
		if string(result.Payload.Data) != GCPSecretNotInitialised {
			str := fmt.Sprintf("secret '%s' not expected - set to '%s' to reinitialise", result.Payload.String(), GCPSecretNotInitialised)
			return errors.New(str)
		}

		// initialises the secret to json {}
		cfg := make(map[string][]byte)
		bytes, err := json.Marshal(cfg)
		if err != nil {
			return err
		}

		return s.setSecret(ctx, bytes)
	}

	return nil
}
