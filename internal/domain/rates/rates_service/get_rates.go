package rates_service

import (
	"context"
	"fmt"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

type RateService struct {
	client rates_interface.ExchangeClient
	repo   rates_interface.RateRepository
}

func NewRateService(client rates_interface.ExchangeClient, repo rates_interface.RateRepository) *RateService {
	return &RateService{client: client, repo: repo}
}

func (s *RateService) GetRates(ctx context.Context, method rates_model.CalcMethod) (*rates_model.Rate, error) {
	depth, err := s.client.FetchDepth(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch depth: %w", err)
	}

	ask, err := method.Calculate(depth.Asks)
	if err != nil {
		return nil, fmt.Errorf("calculate ask: %w", err)
	}

	bid, err := method.Calculate(depth.Bids)
	if err != nil {
		return nil, fmt.Errorf("calculate bid: %w", err)
	}

	rate := &rates_model.Rate{
		Ask:       ask,
		Bid:       bid,
		FetchedAt: depth.Timestamp,
	}

	if err := s.repo.Save(ctx, rate); err != nil {
		return nil, fmt.Errorf("save rate: %w", err)
	}

	return rate, nil
}
