package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
)

const agentIDString string = "AgentID_V2"

type AgentInfo struct {
	BLSPrivateKey string `json:"blsPrivateKey"`
	ECPrivateKey  string `json:"ecPrivateKey"`
	WorkspaceID   string `json:"workspaceID"`
	APIKeyID      string `json:"APIKeyID"`
	APIKeySecret  string `json:"APIKeySecret"`
}

type AgentStore interface {
	StoreWriter
	GetAgentInfo() (*AgentInfo, error)
}

type StoreWriter interface {
	SaveAgentInfo(id string, agent *AgentInfo) error
}

type storage struct {
	kv util.KVStore
}

func NewAgentStore(kv util.KVStore) AgentStore {
	return &storage{
		kv: kv,
	}
}

func (s *storage) SaveAgentInfo(id string, agent *AgentInfo) error {
	if len(id) == 0 {
		return errors.New("invalid agentID")
	}

	if agent == nil {
		return errors.New("invalid agent info")
	}

	data, err := json.Marshal(agent)
	if err != nil {
		return fmt.Errorf("failed to marshal agent info, err: %v", err)
	}

	if err = s.kv.Set(id, data); err != nil {
		return fmt.Errorf("failed to save agent info, err: %v", err)
	}

	if err := s.kv.Set(agentIDString, []byte(id)); err != nil {
		return fmt.Errorf("failed to set agentID, err: %v", err)
	}

	return nil
}

// GetAgentInfo returns the agent info if agent registered, otherwise null. If retrieving from store fails it returns error
func (s storage) GetAgentInfo() (*AgentInfo, error) {
	id, err := s.getSystemAgentID()
	if err != nil {
		if err == defs.ErrKVNotFound {
			// agentID not set, agent not registered
			return nil, nil
		}

		// kv error
		return nil, err
	}

	data, err := s.kv.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve agent info, err: %v", err)
	}

	agentInfo := &AgentInfo{}
	if err = json.Unmarshal(data, agentInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent info, err: %v", err)
	}

	return agentInfo, nil
}

func (s storage) getSystemAgentID() (string, error) {
	d, err := s.kv.Get(agentIDString)
	if err != nil {
		return defs.EmptyString, err
	}

	return bytes.NewBuffer(d).String(), nil
}
