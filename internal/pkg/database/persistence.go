package database

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/metrics"
)

var _ rates_interface.AsyncRatePersister = (*PersistenceWorker)(nil)

type PersistenceWorker struct {
	ch         chan *rates_model.Rate
	repo       rates_interface.RateRepository
	logger     *zap.Logger
	retryMax   int
	retryDelay time.Duration
	done       chan struct{}
	metrics    *metrics.Metrics
}

func NewPersistenceWorker(
	repo rates_interface.RateRepository,
	cfg *config.Config,
	logger *zap.Logger,
	m *metrics.Metrics,
) *PersistenceWorker {
	return &PersistenceWorker{
		ch:         make(chan *rates_model.Rate, cfg.Persist.QueueSize),
		repo:       repo,
		logger:     logger.Named("persistence"),
		retryMax:   cfg.Persist.RetryMax,
		retryDelay: cfg.Persist.RetryDelay,
		done:       make(chan struct{}),
		metrics:    m,
	}
}

func (w *PersistenceWorker) Enqueue(rate *rates_model.Rate) {
	select {
	case w.ch <- rate:
	default:
		// Queue full — drop oldest to make room
		<-w.ch
		w.ch <- rate
		w.logger.Warn("persistence queue full, dropped oldest entry")
	}
	if w.metrics != nil {
		w.metrics.PersistQueueDepth.Set(float64(len(w.ch)))
	}
}

func (w *PersistenceWorker) Start() {
	go func() {
		defer close(w.done)
		for rate := range w.ch {
			w.saveWithRetry(rate)
			if w.metrics != nil {
				w.metrics.PersistQueueDepth.Set(float64(len(w.ch)))
			}
		}
	}()
}

func (w *PersistenceWorker) Close(ctx context.Context) error {
	close(w.ch)

	select {
	case <-w.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *PersistenceWorker) saveWithRetry(rate *rates_model.Rate) {
	delay := w.retryDelay
	for attempt := range w.retryMax {
		start := time.Now()
		if err := w.repo.Save(context.Background(), rate); err != nil {
			if w.metrics != nil {
				w.metrics.DBPersistDuration.Observe(time.Since(start).Seconds())
				if attempt < w.retryMax-1 {
					w.metrics.DBPersistTotal.WithLabelValues("retry").Inc()
				} else {
					w.metrics.DBPersistTotal.WithLabelValues("error").Inc()
				}
			}
			w.logger.Warn("failed to persist rate",
				zap.Int("attempt", attempt+1),
				zap.Int("max", w.retryMax),
				zap.Error(err),
			)
			if attempt < w.retryMax-1 {
				time.Sleep(delay)
				delay *= 2
			}
			continue
		}
		if w.metrics != nil {
			w.metrics.DBPersistDuration.Observe(time.Since(start).Seconds())
			w.metrics.DBPersistTotal.WithLabelValues("ok").Inc()
		}
		return
	}
	w.logger.Error("dropping rate after max retries",
		zap.Float64("ask", float64(rate.Ask)),
		zap.Float64("bid", float64(rate.Bid)),
	)
}
