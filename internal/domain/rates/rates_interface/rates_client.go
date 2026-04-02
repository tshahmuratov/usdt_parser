package rates_interface

import (
	"context"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

type ExchangeClient interface {
	FetchDepth(ctx context.Context) (*rates_model.SpotDepth, error)
}
