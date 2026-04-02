package logger

import (
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	if cfg.Logger.Dev {
		return zap.NewDevelopment()
	}

	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(parseLevel(cfg.Logger.Level))

	return zapCfg.Build()
}

func parseLevel(level string) zapcore.Level {
	var l zapcore.Level
	_ = l.UnmarshalText([]byte(level))
	return l
}
