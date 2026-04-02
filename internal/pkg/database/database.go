package database

import (
	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver for database/sql
	"github.com/jmoiron/sqlx"
	"github.com/tshahmuratov/usdt_parser/internal/pkg/config"
)

func NewDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.Database.DSN())
	if err != nil {
		return nil, err
	}
	return db, nil
}
