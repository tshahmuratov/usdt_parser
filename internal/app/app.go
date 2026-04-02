package app

import (
	"github.com/tshahmuratov/usdt_parser/internal/app/grpcserver"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/database"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/grinex"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/logger"
	"go.uber.org/fx"
)

func New() *fx.App {
	return fx.New(
		config.Module,
		logger.Module,
		database.Module,
		grinex.Module,
		rates.Module,
		grpcserver.Module,
	)
}
