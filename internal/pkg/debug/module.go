package debug

import (
	"context"

	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("debug",
	fx.Provide(NewServer),
	fx.Invoke(func(lc fx.Lifecycle, srv *Server, cfg *config.Config, logger *zap.Logger) {
		if cfg.Debug.Port == 0 {
			logger.Info("debug/pprof server disabled")
			return
		}
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				logger.Info("starting pprof debug server", zap.String("addr", srv.httpSrv.Addr))
				return srv.Start()
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("stopping pprof debug server")
				return srv.Stop(ctx)
			},
		})
	}),
)
