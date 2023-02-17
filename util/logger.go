package util

import (
	"go.uber.org/zap"

	"github.com/qredo/signing-agent/config"
)

func NewLogger(cfg *config.Logging) *zap.SugaredLogger {
	logConfig := zap.NewProductionConfig()

	switch cfg.Format {
	case "text":
		logConfig = zap.NewDevelopmentConfig()
	case "json":
		fallthrough
	default:
		logConfig = zap.NewProductionConfig()
	}

	switch cfg.Level {
	case "info":
		logConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		logConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		logConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		logConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logConfig.DisableStacktrace = true
	l, _ := logConfig.Build()

	return l.Sugar()
}

func NewTestLogger() *zap.SugaredLogger {
	testConfig := &config.Logging{
		Format: "text",
		Level:  "info",
	}
	return NewLogger(testConfig)
}
