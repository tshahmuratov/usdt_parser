package database

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("database",
	fx.Provide(NewDB, NewRateRepo),
	fx.Invoke(func(lc fx.Lifecycle, db *sqlx.DB, logger *zap.Logger) {
		lc.Append(fx.Hook{
			OnStop: func(_ context.Context) error {
				logger.Info("closing database connection")
				return db.Close()
			},
		})
	}),
)
