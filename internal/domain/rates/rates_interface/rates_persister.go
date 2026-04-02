package rates_interface

import (
	"context"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

type AsyncRatePersister interface {
	Enqueue(rate *rates_model.Rate)
	Close(ctx context.Context) error
}
