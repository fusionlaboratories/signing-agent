package util

import (
	"github.com/qredo/signing-agent/config"
)

func CreateStore(cfg *config.Config) KVStore {
	switch cfg.Store.Type {
	case "file":
		return NewFileStore(cfg.Store.FileConfig)
	case "oci":
		return NewOciStore(cfg.Store.OciConfig)
	case "aws":
		return NewAWSStore(cfg.Store.AwsConfig)
	case "gcp":
		return NewGCPStore(cfg.Store.GcpConfig)
	default:
		return nil
	}
}
