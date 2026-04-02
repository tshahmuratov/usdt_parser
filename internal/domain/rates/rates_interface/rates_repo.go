package rates_interface

import (
	"context"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

type RateRepository interface {
	Save(ctx context.Context, rate *rates_model.Rate) error
}
