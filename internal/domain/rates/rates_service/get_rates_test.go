package rates_service_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_service"
	mocks "github.com/tshahmuratov/usdt_parser/mocks/domain/rates/rates_interface"
)

// countingClient wraps MockExchangeClient and counts FetchDepth invocations atomically.
type countingClient struct {
	mocks.MockExchangeClient
	calls atomic.Int32
}

func (c *countingClient) FetchDepth(ctx context.Context) (*rates_model.SpotDepth, error) {
	c.calls.Add(1)
	return c.MockExchangeClient.FetchDepth(ctx)
}

func TestRateService_GetRates_Singleflight(t *testing.T) {
	now := time.Now().UTC()
	depth := &rates_model.SpotDepth{
		Asks:      entries(80, 81, 82),
		Bids:      entries(79, 78, 77),
		Timestamp: now,
	}

	client := new(countingClient)
	persister := mocks.NewMockAsyncRatePersister(t)
	svc := rates_service.NewRateService(client, persister, nil)

	// FetchDepth blocks briefly to allow goroutines to coalesce
	client.On("FetchDepth", mock.Anything).Run(func(_ mock.Arguments) {
		time.Sleep(50 * time.Millisecond)
	}).Return(depth, nil)
	persister.On("Enqueue", mock.Anything).Return()

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make([]error, n)

	for i := range n {
		go func() {
			defer wg.Done()
			_, errs[i] = svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "goroutine %d", i)
	}
	// Singleflight should coalesce all concurrent calls into 1 FetchDepth invocation
	assert.Equal(t, int32(1), client.calls.Load(), "expected exactly 1 FetchDepth call")
}

func TestRateService_GetRates(t *testing.T) {
	now := time.Now().UTC()
	depth := &rates_model.SpotDepth{
		Asks:      entries(80, 81, 82),
		Bids:      entries(79, 78, 77),
		Timestamp: now,
	}

	t.Run("success with TopN", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)
		persister.On("Enqueue", mock.Anything).Return()

		method := rates_model.TopN{N: 0}
		rate, err := svc.GetRates(context.Background(), method)

		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(80), rate.Ask)
		assert.Equal(t, rates_model.Price(79), rate.Bid)
		assert.Equal(t, now, rate.FetchedAt)
	})

	t.Run("client error", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		client.On("FetchDepth", mock.Anything).Return(nil, rates_model.ErrFetchFailed)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.ErrorIs(t, err, rates_model.ErrFetchFailed)
	})

	t.Run("enqueue is called", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)
		persister.On("Enqueue", mock.Anything).Return()

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		persister.AssertCalled(t, "Enqueue", mock.Anything)
	})

	t.Run("calc error", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		client.On("FetchDepth", mock.Anything).Return(depth, nil)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 99})
		require.ErrorIs(t, err, rates_model.ErrIndexOutOfBounds)
	})

	t.Run("fallback on fetch error after success", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		// First call succeeds — populates fallback
		client.On("FetchDepth", mock.Anything).Return(depth, nil).Once()
		persister.On("Enqueue", mock.Anything).Return()

		rate, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(80), rate.Ask)

		// Second call fails — should use fallback
		client.On("FetchDepth", mock.Anything).Return(nil, rates_model.ErrFetchFailed).Once()

		rate, err = svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(80), rate.Ask)
		assert.Equal(t, now, rate.FetchedAt)
	})

	t.Run("cold start fetch error returns error", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)

		client.On("FetchDepth", mock.Anything).Return(nil, rates_model.ErrFetchFailed)

		_, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.ErrorIs(t, err, rates_model.ErrFetchFailed)
	})

	t.Run("successful fetch updates fallback", func(t *testing.T) {
		client := mocks.NewMockExchangeClient(t)
		persister := mocks.NewMockAsyncRatePersister(t)
		svc := rates_service.NewRateService(client, persister, nil)
		persister.On("Enqueue", mock.Anything).Return()

		depth1 := &rates_model.SpotDepth{
			Asks:      entries(80, 81, 82),
			Bids:      entries(79, 78, 77),
			Timestamp: now,
		}
		later := now.Add(time.Second)
		depth2 := &rates_model.SpotDepth{
			Asks:      entries(90, 91, 92),
			Bids:      entries(89, 88, 87),
			Timestamp: later,
		}

		// First call with depth1
		client.On("FetchDepth", mock.Anything).Return(depth1, nil).Once()
		rate, err := svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(80), rate.Ask)

		// Second call with depth2 — fallback should be updated
		client.On("FetchDepth", mock.Anything).Return(depth2, nil).Once()
		rate, err = svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(90), rate.Ask)

		// Third call fails — should use depth2 (latest fallback)
		client.On("FetchDepth", mock.Anything).Return(nil, rates_model.ErrFetchFailed).Once()
		rate, err = svc.GetRates(context.Background(), rates_model.TopN{N: 0})
		require.NoError(t, err)
		assert.Equal(t, rates_model.Price(90), rate.Ask)
		assert.Equal(t, later, rate.FetchedAt)
	})
}
