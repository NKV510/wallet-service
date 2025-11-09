package repository

import (
	"context"
	"testing"

	"github.com/NKV510/wallet-service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbPool, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost:5433/wallet_test")
	require.NoError(t, err)

	_, err = dbPool.Exec(context.Background(), "DELETE FROM wallets")
	require.NoError(t, err)

	return dbPool
}

func TestWalletRepository_CreateWallet(t *testing.T) {
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	repo := NewWalletRepository(dbPool)

	tests := []struct {
		name      string
		walletID  string
		wantError bool
	}{
		{
			name:      "successful wallet creation",
			walletID:  "123e4567-e89b-12d3-a456-426614174000",
			wantError: false,
		},
		{
			name:      "duplicate wallet creation",
			walletID:  "123e4567-e89b-12d3-a456-426614174000",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateWallet(context.Background(), tt.walletID)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWalletRepository_GetWallet(t *testing.T) {
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	repo := NewWalletRepository(dbPool)

	walletID := "123e4567-e89b-12d3-a456-426614174000"
	err := repo.CreateWallet(context.Background(), walletID)
	require.NoError(t, err)

	tests := []struct {
		name      string
		walletID  string
		wantError bool
	}{
		{
			name:      "get existing wallet",
			walletID:  walletID,
			wantError: false,
		},
		{
			name:      "get non-existing wallet",
			walletID:  "00000000-0000-0000-0000-000000000000", // Валидный UUID
			wantError: true,                                   // Должна быть ошибка!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := repo.GetWallet(context.Background(), tt.walletID)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, walletID, wallet.ID)
				assert.Equal(t, int64(0), wallet.Balance)
			}
		})
	}
}

func TestWalletRepository_UpdateWalletBalance(t *testing.T) {
	dbPool := setupTestDB(t)
	defer dbPool.Close()

	repo := NewWalletRepository(dbPool)

	// Создаем тестовый кошелек
	walletID := "123e4567-e89b-12d3-a456-426614174000"
	err := repo.CreateWallet(context.Background(), walletID)
	require.NoError(t, err)

	tests := []struct {
		name          string
		walletID      string
		operationType models.OperationType
		amount        int64
		wantError     bool
		errorContains string
	}{
		{
			name:          "successful deposit",
			walletID:      walletID,
			operationType: models.DEPOSIT,
			amount:        1000,
			wantError:     false,
		},
		{
			name:          "successful withdraw",
			walletID:      walletID,
			operationType: models.WITHDRAW,
			amount:        500,
			wantError:     false,
		},
		{
			name:          "insufficient funds",
			walletID:      walletID,
			operationType: models.WITHDRAW,
			amount:        1000,
			wantError:     true,
			errorContains: "insufficient funds",
		},
		{
			name:          "invalid operation type",
			walletID:      walletID,
			operationType: "INVALID",
			amount:        100,
			wantError:     true,
			errorContains: "invalid operation type",
		},
		{
			name:          "non-existing wallet",
			walletID:      "00000000-0000-0000-0000-000000000000", // Валидный UUID
			operationType: models.DEPOSIT,
			amount:        100,
			wantError:     true,
			errorContains: "wallet not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateWalletBalance(context.Background(), tt.walletID, tt.operationType, tt.amount)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)

				// Проверяем, что баланс обновился корректно
				wallet, err := repo.GetWallet(context.Background(), walletID)
				assert.NoError(t, err)
				assert.NotNil(t, wallet)

				if tt.operationType == models.DEPOSIT {
					assert.True(t, wallet.Balance >= tt.amount)
				} else if tt.operationType == models.WITHDRAW {
					assert.True(t, wallet.Balance >= 0)
				}
			}
		})
	}
}
