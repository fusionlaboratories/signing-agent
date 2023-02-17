package lib

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/qredo/signing-agent/defs"
	"github.com/qredo/signing-agent/util"
)

var agentIDString = "AgentID"

type Storage struct {
	kv util.KVStore
}

func NewStore(store util.KVStore) *Storage {
	s := &Storage{
		kv: store,
	}

	return s
}

type Agent struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	BLSSeed     []byte `json:"bls_seed"`
	AccountCode string `json:"account_code,omitempty"`
	ZKPID       []byte `json:"zkpid,omitempty"`
	ZKPToken    []byte `json:"zkptoken,omitempty"`
	Pending     bool   `json:"pending"`
}

func (s *Storage) AddPending(ref string, c *Agent) error {
	c.Pending = true
	_, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if err = s.kv.Set(ref, data); err != nil {
		return err
	}

	return nil
}

func (s *Storage) RemovePending(ref string) error {
	d, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if !c.Pending {
		return errors.New("agent not pending")
	}

	return s.kv.Del(ref)
}

func (s *Storage) GetPending(ref string) *Agent {
	d, err := s.kv.Get(ref)
	if err != nil && err != defs.KVErrNotFound {
		return nil
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if !c.Pending {
		return nil
	}

	return c
}

func (s *Storage) AddAgent(id string, c *Agent) error {
	_, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c.Pending = false
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	if err = s.kv.Set(id, data); err != nil {
		return err
	}

	return nil
}

func (s *Storage) RemoveAgent(id string) error {
	d, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return err
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return err
	}

	if c.Pending {
		return errors.New("agent pending")
	}

	return s.kv.Del(id)
}

func (s *Storage) GetAgent(id string) *Agent {
	d, err := s.kv.Get(id)
	if err != nil && err != defs.KVErrNotFound {
		return nil
	}

	c := &Agent{}
	err = json.Unmarshal(d, c)
	if err != nil {
		return nil
	}

	if c.Pending {
		return nil
	}

	return c
}

func (s *Storage) GetSystemAgentID() string {
	d, err := s.kv.Get(agentIDString)
	if err != nil && err != defs.KVErrNotFound {
		return ""
	}
	return bytes.NewBuffer(d).String()
}

func (s *Storage) SetSystemAgentID(agentID string) error {
	if err := s.kv.Set(agentIDString, []byte(agentID)); err != nil {
		return err
	}
	return nil
}
