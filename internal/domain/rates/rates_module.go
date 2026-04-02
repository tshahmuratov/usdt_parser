package rates

import (
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_handler"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_service"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/database"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/grinex"
	"go.uber.org/fx"
)

var Module = fx.Module("rates",
	fx.Provide(
		// Bind infrastructure to domain interfaces
		fx.Annotate(
			func(r *database.RateRepo) rates_interface.RateRepository { return r },
			fx.As(new(rates_interface.RateRepository)),
		),
		fx.Annotate(
			func(c *grinex.GrinexClient) rates_interface.ExchangeClient { return c },
			fx.As(new(rates_interface.ExchangeClient)),
		),
		// Domain service
		fx.Annotate(
			rates_service.NewRateService,
			fx.As(new(rates_handler.RateServicer)),
		),
		// Handler adapter
		rates_handler.NewRatesHandler,
	),
)
