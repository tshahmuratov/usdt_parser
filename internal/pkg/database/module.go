package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
)

var Module = fx.Module("database",
	fx.Provide(
		NewDB,
		NewRateRepo,
		NewPersistenceWorker,
		fx.Annotate(
			func(pw *PersistenceWorker) rates_interface.AsyncRatePersister { return pw },
			fx.As(new(rates_interface.AsyncRatePersister)),
		),
	),
	fx.Invoke(func(lc fx.Lifecycle, db *sqlx.DB, pw *PersistenceWorker, logger *zap.Logger) {
		lc.Append(fx.Hook{
			OnStart: func(_ context.Context) error {
				logger.Info("starting persistence worker")
				pw.Start()
				return nil
			},
			OnStop: func(_ context.Context) error {
				logger.Info("draining persistence queue")
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				if err := pw.Close(ctx); err != nil {
					logger.Error("persistence drain timeout", zap.Error(err))
				}
				logger.Info("closing database connection")
				return db.Close()
			},
		})
	}),
)
