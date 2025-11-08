package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBPool(host, port, user, password, dbname string, maxConns int32) (*pgxpool.Pool, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbname)

	dbConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	dbConfig.MaxConns = maxConns
	dbConfig.HealthCheckPeriod = 1 * time.Minute
	dbConfig.MaxConnLifetime = 1 * time.Hour
	dbConfig.MaxConnIdleTime = 30 * time.Minute

	var dbPool *pgxpool.Pool
	maxRetries := 10
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		log.Printf("Attempt %d/%d: Connecting to database at %s:%s", i+1, maxRetries, host, port)

		dbPool, err = pgxpool.NewWithConfig(ctx, dbConfig)
		if err != nil {
			cancel()
			log.Printf("Attempt %d/%d: Failed to connect to database: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				log.Printf("Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("unable to connect to database after %d attempts: %w", maxRetries, err)
		}

		if err := dbPool.Ping(ctx); err != nil {
			cancel()
			log.Printf("Attempt %d/%d: Database ping failed: %v", i+1, maxRetries, err)
			dbPool.Close()
			if i < maxRetries-1 {
				log.Printf("Retrying in %v...", retryDelay)
				time.Sleep(retryDelay)
				continue
			}
			return nil, fmt.Errorf("unable to ping database after %d attempts: %w", maxRetries, err)
		}

		cancel()
		log.Printf("Successfully connected to database on attempt %d", i+1)
		break
	}

	if err := checkWalletsTable(context.Background(), dbPool); err != nil {
		dbPool.Close()
		return nil, fmt.Errorf("wallets table check failed: %w", err)
	}

	return dbPool, nil
}

func checkWalletsTable(ctx context.Context, dbPool *pgxpool.Pool) error {
	var tableExists bool
	err := dbPool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'wallets'
		)
	`).Scan(&tableExists)

	if err != nil {
		return fmt.Errorf("failed to check wallets table: %w", err)
	}

	if !tableExists {
		return fmt.Errorf("wallets table does not exist. Please run migrations first")
	}

	log.Println("Wallets table exists and is ready")
	return nil
}
