package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NKV510/wallet-service/internal/models"
	"github.com/NKV510/wallet-service/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*WalletHandler, *pgxpool.Pool) {
	t.Helper()

	dbPool, err := pgxpool.New(context.Background(), "postgres://postgres:password@localhost:5433/wallet_test")
	require.NoError(t, err)

	_, err = dbPool.Exec(context.Background(), "DELETE FROM wallets")
	require.NoError(t, err)

	repo := repository.NewWalletRepository(dbPool)
	handler := NewWalletHandler(repo)

	return handler, dbPool
}

func TestWalletHandler_ProcessOperation(t *testing.T) {
	handler, dbPool := setupTestHandler(t)
	defer dbPool.Close()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful deposit",
			requestBody: models.WalletOperation{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: models.DEPOSIT,
				Amount:        1000,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "successful withdraw",
			requestBody: models.WalletOperation{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: models.WITHDRAW,
				Amount:        500,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid operation type",
			requestBody: models.WalletOperation{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: "INVALID",
				Amount:        100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid operation type",
		},
		{
			name: "negative amount",
			requestBody: models.WalletOperation{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: models.DEPOSIT,
				Amount:        -100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "amount must be positive",
		},
		{
			name: "zero amount",
			requestBody: models.WalletOperation{
				WalletID:      "123e4567-e89b-12d3-a456-426614174000",
				OperationType: models.DEPOSIT,
				Amount:        0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "amount must be positive",
		},
		{
			name: "empty wallet ID",
			requestBody: models.WalletOperation{
				WalletID:      "",
				OperationType: models.DEPOSIT,
				Amount:        100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "wallet ID is required",
		},
		{
			name:           "invalid json",
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req, err := http.NewRequest("POST", "/api/v1/wallet", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ProcessOperation(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "success", response["status"])
			}
		})
	}
}

func TestWalletHandler_GetWalletBalance(t *testing.T) {
	handler, dbPool := setupTestHandler(t)
	defer dbPool.Close()

	walletID := "123e4567-e89b-12d3-a456-426614174000"
	repo := repository.NewWalletRepository(dbPool)
	err := repo.CreateWallet(context.Background(), walletID)
	require.NoError(t, err)
	err = repo.UpdateWalletBalance(context.Background(), walletID, models.DEPOSIT, 1500)
	require.NoError(t, err)

	tests := []struct {
		name           string
		walletID       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "get existing wallet balance",
			walletID:       walletID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existing wallet balance",
			walletID:       "00000000-0000-0000-0000-000000000000",
			expectedStatus: http.StatusNotFound,
			expectedError:  "wallet not found",
		},
		{
			name:           "empty wallet id",
			walletID:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Wallet ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/wallets/"+tt.walletID, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{gin.Param{Key: "walletId", Value: tt.walletID}}

			handler.GetWalletBalance(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			} else if tt.expectedStatus == http.StatusOK {
				var response models.WalletBalanceResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, walletID, response.WalletID)
				assert.Equal(t, int64(1500), response.Balance)
			}
		})
	}
}
