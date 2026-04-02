package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_interface"
	"github.com/tshahmuratov/usdt_parser/internal/domain/rates/rates_model"
)

var _ rates_interface.RateRepository = (*RateRepo)(nil)

type RateRepo struct {
	db *sqlx.DB
}

func NewRateRepo(db *sqlx.DB) *RateRepo {
	return &RateRepo{db: db}
}

func (r *RateRepo) Save(ctx context.Context, rate *rates_model.Rate) error {
	const query = `INSERT INTO rates (ask, bid, fetched_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	row := r.db.QueryRowxContext(ctx, query,
		float64(rate.Ask), float64(rate.Bid), rate.FetchedAt,
	)

	var id int64
	var createdAt time.Time
	if err := row.Scan(&id, &createdAt); err != nil {
		return fmt.Errorf("%w: %v", rates_model.ErrStoreFailed, err)
	}

	rate.ID = id
	rate.CreatedAt = createdAt
	return nil
}
