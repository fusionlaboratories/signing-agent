package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/qredo/signing-agent/config"
	"github.com/qredo/signing-agent/util"
)

const (
	TestDataDBStoreFilePath = "../testdata/test-store.db"
)

func NewMock(cfg *config.Config, kv util.KVStore) (*signingAgent, error) {
	return &signingAgent{
		cfg:   cfg,
		store: NewStore(kv),
		htc:   util.NewHTTPMockClient(),
	}, nil
}

func TestCreateSigningAgentClient(t *testing.T) {
	t.Run(
		"Create a signingAgent",
		func(t *testing.T) {
			var (
				cfg *config.Config
				kv  util.KVStore
				err error
			)
			cfg = &config.Config{
				Base: config.Base{
					PIN:      1234,
					QredoAPI: "https://play-api.qredo.network/api/v1/p",
				},
				AutoApprove: config.AutoApprove{
					Enabled: true,
				},
			}

			kv = util.NewFileStore(TestDataDBStoreFilePath)
			err = kv.Init()
			require.Nil(t, err)

			core, err := NewMock(cfg, kv)
			assert.NoError(t, err)
			assert.Equal(t, core.cfg.Base.PIN, cfg.Base.PIN)
		})
}
