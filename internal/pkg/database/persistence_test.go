package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	mocks "github.com/tshahmuratov/usdt_parser/mocks/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/database"
)

func newTestWorker(repo *mocks.MockRateRepository, queueSize int) *database.PersistenceWorker {
	cfg := &config.Config{
		Persist: config.PersistConfig{
			QueueSize:  queueSize,
			RetryMax:   3,
			RetryDelay: 1 * time.Millisecond,
		},
	}
	return database.NewPersistenceWorker(repo, cfg, zap.NewNop(), nil)
}

func TestPersistenceWorker_EnqueueAndSave(t *testing.T) {
	repo := mocks.NewMockRateRepository(t)
	w := newTestWorker(repo, 10)

	rate := &rates_model.Rate{Ask: 80, Bid: 79, FetchedAt: time.Now()}
	repo.On("Save", mock.Anything, rate).Return(nil)

	w.Start()
	w.Enqueue(rate)

	// Allow time for consumer to process
	time.Sleep(50 * time.Millisecond)

	err := w.Close(context.Background())
	require.NoError(t, err)

	repo.AssertCalled(t, "Save", mock.Anything, rate)
}

func TestPersistenceWorker_OverflowDropsOldest(t *testing.T) {
	repo := mocks.NewMockRateRepository(t)
	w := newTestWorker(repo, 2)

	rate1 := &rates_model.Rate{Ask: 1, Bid: 1, FetchedAt: time.Now()}
	rate2 := &rates_model.Rate{Ask: 2, Bid: 2, FetchedAt: time.Now()}
	rate3 := &rates_model.Rate{Ask: 3, Bid: 3, FetchedAt: time.Now()}

	// Fill the queue without starting the worker (so items stay in channel)
	w.Enqueue(rate1)
	w.Enqueue(rate2)
	// This should drop rate1 (oldest) and add rate3
	w.Enqueue(rate3)

	saved := make([]*rates_model.Rate, 0)
	repo.On("Save", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		saved = append(saved, args.Get(1).(*rates_model.Rate))
	}).Return(nil)

	w.Start()
	err := w.Close(context.Background())
	require.NoError(t, err)

	// rate1 was dropped; rate2 and rate3 should be saved
	require.Len(t, saved, 2)
	assert.Equal(t, rates_model.Price(2), saved[0].Ask)
	assert.Equal(t, rates_model.Price(3), saved[1].Ask)
}

func TestPersistenceWorker_RetryOnSaveFailure(t *testing.T) {
	repo := mocks.NewMockRateRepository(t)
	w := newTestWorker(repo, 10)

	rate := &rates_model.Rate{Ask: 80, Bid: 79, FetchedAt: time.Now()}

	// Fail twice, succeed on third attempt
	repo.On("Save", mock.Anything, rate).Return(rates_model.ErrStoreFailed).Twice()
	repo.On("Save", mock.Anything, rate).Return(nil).Once()

	w.Start()
	w.Enqueue(rate)

	time.Sleep(100 * time.Millisecond)

	err := w.Close(context.Background())
	require.NoError(t, err)

	repo.AssertNumberOfCalls(t, "Save", 3)
}

func TestPersistenceWorker_GracefulDrain(t *testing.T) {
	repo := mocks.NewMockRateRepository(t)
	w := newTestWorker(repo, 100)

	repo.On("Save", mock.Anything, mock.Anything).Return(nil)

	// Enqueue multiple rates before starting
	for i := range 5 {
		w.Enqueue(&rates_model.Rate{
			Ask:       rates_model.Price(i),
			Bid:       rates_model.Price(i),
			FetchedAt: time.Now(),
		})
	}

	w.Start()
	err := w.Close(context.Background())
	require.NoError(t, err)

	// All 5 should have been saved during drain
	repo.AssertNumberOfCalls(t, "Save", 5)
}
