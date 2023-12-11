package action

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/qredo/signing-agent/crypto"
	"github.com/qredo/signing-agent/internal/auth"
	"github.com/qredo/signing-agent/internal/defs"
	"github.com/qredo/signing-agent/internal/util"
	"go.uber.org/zap"
)

type signRequest struct {
	Status     int      `json:"status"`
	Signatures []string `json:"signatures"`
}

type getActionResponse struct {
	Messages []string `json:"messages"`
	Status   int      `json:"status"`
	ID       string   `json:"id"`
}

const (
	approve int = 3
	reject  int = 4
)

type Signer interface {
	SetKey(blsPrivateKey string) error
	ActionApprove(actionID string) error
	ActionReject(actionID string) error
	ApproveActionMessage(actionID string, message []byte) error
}

type actionSigner struct {
	baseURL       string
	blsPrivateKey []byte

	htc          *util.Client
	authProvider auth.HeaderProvider
	log          *zap.SugaredLogger
}

func NewSigner(baseURL string, authProvide auth.HeaderProvider, log *zap.SugaredLogger, blsPrivateKey string) (Signer, error) {
	s := &actionSigner{
		htc:          util.NewHTTPClient(),
		authProvider: authProvide,
		baseURL:      baseURL,
		log:          log,
	}

	if len(blsPrivateKey) > 0 {
		if err := s.SetKey(blsPrivateKey); err != nil {
			return nil, fmt.Errorf("invalid bls key")
		}
	}

	return s, nil
}

func (s *actionSigner) SetKey(blsPrivateKey string) error {
	data, err := base64.StdEncoding.DecodeString(blsPrivateKey)
	if err != nil {
		s.log.Errorf("failed to decode the agent blsPrivateKey, err: %v", err)
		return fmt.Errorf("invalid bls key")
	}

	s.blsPrivateKey = data
	return nil
}

func (s actionSigner) ActionApprove(actionID string) error {
	message, err := s.getActionMessage(actionID)
	if err != nil {
		return err
	}

	return s.signAction(actionID, message, approve)
}

func (s actionSigner) ActionReject(actionID string) error {
	message, err := s.getActionMessage(actionID)
	if err != nil {
		return err
	}

	return s.signAction(actionID, message, reject)
}

func (s actionSigner) ApproveActionMessage(actionID string, message []byte) error {
	return s.signAction(actionID, message, approve)
}

func (s actionSigner) getActionMessage(actionID string) ([]byte, error) {
	resp := &getActionResponse{}

	header := s.authProvider.GetAuthHeader()
	if err := s.htc.Request(http.MethodGet, defs.URLAction(s.baseURL, actionID), nil, resp, header); err != nil {
		return nil, err
	}

	if resp.Status != defs.StatusPending {
		return nil, defs.ErrBadRequest().WithDetail("action can't be signed, status not pending")
	}

	message, err := hex.DecodeString(resp.Messages[0])
	if err != nil {
		s.log.Errorf("failed to decode the action message, err: %v", err)
		return nil, defs.ErrInternal().WithDetail("failed to decode the action message")
	}
	return message, nil
}

func (s actionSigner) signAction(actionID string, message []byte, status int) error {
	if len(s.blsPrivateKey) == 0 {
		return defs.ErrInternal().WithDetail("failed to generate signature, invalid blsKey")
	}

	blsSig, _ := crypto.BLSSign(message, s.blsPrivateKey)
	signature := hex.EncodeToString(blsSig)

	req := signRequest{
		Status:     status,
		Signatures: []string{signature},
	}

	header := s.authProvider.GetAuthHeader()
	if err := s.htc.Request(http.MethodPost, defs.URLAction(s.baseURL, actionID), req, nil, header); err != nil {
		s.log.Errorf("error while signing the action `%s`, err:%v", actionID, err)
		return defs.ErrInternal().WithDetail("failed to sign the action")
	}

	return nil
}
