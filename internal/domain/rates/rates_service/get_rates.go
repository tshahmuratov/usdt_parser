package rates_service

import (
	"context"
	"fmt"
	"sync/atomic"

	"golang.org/x/sync/singleflight"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/metrics"
)

type RateService struct {
	client    rates_interface.ExchangeClient
	persister rates_interface.AsyncRatePersister
	sfg       singleflight.Group
	lastDepth atomic.Pointer[rates_model.SpotDepth]
	metrics   *metrics.Metrics
}

func NewRateService(client rates_interface.ExchangeClient, persister rates_interface.AsyncRatePersister, m *metrics.Metrics) *RateService {
	return &RateService{client: client, persister: persister, metrics: m}
}

func (s *RateService) GetRates(ctx context.Context, method rates_model.CalcMethod) (*rates_model.Rate, error) {
	depth, err := s.fetchDepth(ctx)
	if err != nil {
		return nil, err
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

	s.persister.Enqueue(rate)

	return rate, nil
}

func (s *RateService) fetchDepth(ctx context.Context) (*rates_model.SpotDepth, error) {
	v, err, shared := s.sfg.Do("fetch_depth", func() (interface{}, error) {
		return s.client.FetchDepth(ctx)
	})

	if s.metrics != nil {
		s.metrics.SingleflightTotal.Inc()
		if shared {
			s.metrics.SingleflightShared.Inc()
		}
	}

	if err != nil {
		// Fallback to last known depth
		if cached := s.lastDepth.Load(); cached != nil {
			if s.metrics != nil {
				s.metrics.FallbackTotal.Inc()
			}
			return cached, nil
		}
		return nil, fmt.Errorf("fetch depth: %w", err)
	}

	depth := v.(*rates_model.SpotDepth)
	s.lastDepth.Store(depth)

	return depth, nil
}
