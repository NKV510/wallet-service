package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func PoolConection(DATABASE_URL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), DATABASE_URL)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
