package rates_service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_service"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) FetchDepth(ctx context.Context) (*rates_model.SpotDepth, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rates_model.SpotDepth), args.Error(1)
}

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Save(ctx context.Context, rate *rates_model.Rate) error {
	args := m.Called(ctx, rate)
	return args.Error(0)
}

func TestRateService_GetRates(t *testing.T) {
	now := time.Now().UTC()
	depth := &rates_model.SpotDepth{
		Asks:      entries(80, 81, 82),
		Bids:      entries(79, 78, 77),
		Timestamp: now,
	}

	t.Run("success with TopN", func(t *testing.T) {
		client := new(mockClient)
		repo := new(mockRepo)
		svc := rates_service.NewRateService(client, repo)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(nil)

		method := rates_model.TopN{N: 0}
		rate, err := svc.GetRates(context.Background(), method)

		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(80), rate.Ask)
		assert.Equal(t, rates_model.Price(79), rate.Bid)
		assert.Equal(t, now, rate.FetchedAt)
		client.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("client error", func(t *testing.T) {
		client := new(mockClient)
		repo := new(mockRepo)
		svc := rates_service.NewRateService(client, repo)

		client.On("FetchDepth", mock.Anything).Return(nil, rates_model.ErrFetchFailed)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.ErrorIs(t, err, rates_model.ErrFetchFailed)
	})

	t.Run("save error", func(t *testing.T) {
		client := new(mockClient)
		repo := new(mockRepo)
		svc := rates_service.NewRateService(client, repo)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)
		repo.On("Save", mock.Anything, mock.Anything).Return(rates_model.ErrStoreFailed)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.ErrorIs(t, err, rates_model.ErrStoreFailed)
	})

	t.Run("calc error", func(t *testing.T) {
		client := new(mockClient)
		repo := new(mockRepo)
		svc := rates_service.NewRateService(client, repo)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 99})
		require.ErrorIs(t, err, rates_model.ErrIndexOutOfBounds)
	})
}
