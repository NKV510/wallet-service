package repository

import (
	"context"
	"fmt"

	"github.com/NKV510/wallet-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepository struct {
	db *pgxpool.Pool
}

func NewWalletRepository(db *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) CreateWallet(ctx context.Context, walletID string) error {
	query := `INSERT INTO wallets (id, balance) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, walletID, 0)
	return err
}

func (r *WalletRepository) GetWallet(ctx context.Context, walletID string) (*models.Wallet, error) {
	query := `SELECT id, balance, created_at, updated_at FROM wallets WHERE id = $1`

	var wallet models.Wallet
	err := r.db.QueryRow(ctx, query, walletID).Scan(
		&wallet.ID,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &wallet, nil
}

func (r *WalletRepository) UpdateWalletBalance(ctx context.Context, walletID string, operationType models.OperationType, amount int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Сначала проверяем существование кошелька и блокируем строку для обновления
	var currentBalance int64
	err = tx.QueryRow(ctx, "SELECT balance FROM wallets WHERE id = $1 FOR UPDATE", walletID).Scan(&currentBalance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("wallet not found")
		}
		return fmt.Errorf("failed to get wallet balance: %w", err)
	}

	// Вычисляем новый баланс
	var newBalance int64
	switch operationType {
	case models.DEPOSIT:
		newBalance = currentBalance + amount
	case models.WITHDRAW:
		newBalance = currentBalance - amount
		if newBalance < 0 {
			return fmt.Errorf("insufficient funds")
		}
	default:
		return fmt.Errorf("invalid operation type")
	}

	// Обновляем баланс
	query := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.Exec(ctx, query, newBalance, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	return tx.Commit(ctx)
}
