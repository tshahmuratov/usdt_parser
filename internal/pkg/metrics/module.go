package metrics

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("metrics",
	fx.Provide(NewMetrics, NewServer),
	fx.Invoke(func(lc fx.Lifecycle, srv *Server, logger *zap.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				logger.Info("starting metrics server", zap.String("addr", srv.httpSrv.Addr))
				return srv.Start()
			},
			OnStop: func(ctx context.Context) error {
				logger.Info("stopping metrics server")
				return srv.Stop(ctx)
			},
		})
	}),
)
