package api

import (
	"errors"
	"strings"
)

const (
	maxBase64FieldSize = 16384
	maxStringFieldSize = 256
)

// swagger:model ClientRegisterRequest
type ClientRegisterRequest struct {
	// The name of the agent
	// example: test-agent
	Name string `json:"name"`

	// The API key for the partner API
	// example: eyJrZXkiOiJHM0Fo... (truncated)
	APIKey string `json:"APIKey"`

	// The Base64-encoded private key pem of which the public key has been registered in the Partner API
	// example: LS0tLS1CRUdJTiBS... (truncated)
	Base64PrivateKey string `json:"base64PrivateKey"`
}

func (r *ClientRegisterRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.APIKey = strings.TrimSpace(r.APIKey)
	r.Base64PrivateKey = strings.TrimSpace(r.Base64PrivateKey)

	switch {
	case r.Name == "" || len(r.Name) > maxStringFieldSize:
		return errors.New("name")
	case r.APIKey == "" || len(r.APIKey) > maxStringFieldSize:
		return errors.New("APIKey")
	case r.Base64PrivateKey == "" || len(r.Base64PrivateKey) > maxBase64FieldSize:
		return errors.New("base64PrivateKey")
	default:
		return nil
	}
}

// swagger:ignore
type ClientRegisterResponse struct {
	BLSPublicKey string `json:"bls_public_key"`
	ECPublicKey  string `json:"ec_public_key"`
	RefID        string `json:"ref_id"`
}

// swagger:ignore
type ClientRegisterFinishRequest struct {
	ID           string `json:"id"`
	AccountCode  string `json:"accountCode"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	IDDocument   string `json:"idDoc"`
}

// swagger:ignore
type ClientRegisterFinishResponse struct {
	FeedURL string `json:"feed_url"`
}

// swagger:ignore
type CoreClientServiceRegisterFinishRequest struct {
	ClientID          string `json:"client_id,omitempty"`
	IDDocSignatureHex string `json:"idDocSignatureHex"`
}

// swagger:ignore
type CoreClientServiceRegisterFinishResponse struct {
	Feed string `json:"feed"`
}

// swagger:ignore
type QredoRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

func NewQredoRegisterInitRequest(name, blsPublicKey, ecPublicKey string) *QredoRegisterInitRequest {
	return &QredoRegisterInitRequest{
		Name:         name,
		BLSPublicKey: blsPublicKey,
		ECPublicKey:  ecPublicKey,
	}
}

// swagger:ignore
type QredoRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

// swagger:model AgentRegisterResponse
type AgentRegisterResponse struct {
	// The ID of the agent
	// example: 98cTMMSPrDdcDDVU8idhuJGK2U1P4vmQcsp8wnED8pPR
	AgentID string `json:"agentID"`

	// The feed websocket URL
	// example: http://localhost:8007/api/v1/client/feed
	FeedURL string `json:"feedURL"`
}

// swagger:model GetClientResponse
type GetClientResponse struct {
	// The ID of the agent
	// example: 98cTMMSPrDdcDDVU8idhuJGK2U1P4vmQcsp8wnED8pPR
	AgentID string `json:"agentID"`

	// The name of the agent
	// example: some-agent
	AgentName string `json:"agentName"`

	// The feed websocket URL
	// example: http://localhost:8007/api/v1/client/feed
	FeedURL string `json:"feedURL"`
}
