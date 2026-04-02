package grinex

import (
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/metrics"
	"go.uber.org/fx"
)

var Module = fx.Module("grinex",
	fx.Provide(func(cfg *config.Config, m *metrics.Metrics) *GrinexClient {
		return NewGrinexClient(cfg.Grinex.BaseURL, cfg.Grinex.Timeout, cfg.Grinex.DepthLimit, m)
	}),
)
