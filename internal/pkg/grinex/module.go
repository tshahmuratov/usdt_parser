package grinex

import (
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"go.uber.org/fx"
)

var Module = fx.Module("grinex",
	fx.Provide(func(cfg *config.Config) *GrinexClient {
		return NewGrinexClient(cfg.Grinex.BaseURL, cfg.Grinex.Timeout)
	}),
)
