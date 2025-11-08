package handlers

import (
	"net/http"

	"github.com/NKV510/wallet-service/internal/models"
	"github.com/NKV510/wallet-service/internal/repository"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletRepo *repository.WalletRepository
}

func NewWalletHandler(walletRepo *repository.WalletRepository) *WalletHandler {
	return &WalletHandler{walletRepo: walletRepo}
}

func (h *WalletHandler) ProcessOperation(ctx *gin.Context) {
	var operation models.WalletOperation

	if err := ctx.ShouldBindJSON(&operation); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Проверяем валидность операции
	if operation.Amount <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	if operation.OperationType != models.DEPOSIT && operation.OperationType != models.WITHDRAW {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation type"})
		return
	}

	// Проверяем существование кошелька
	wallet, err := h.walletRepo.GetWallet(ctx.Request.Context(), operation.WalletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get wallet"})
		return
	}

	// Если кошелек не существует, создаем его
	if wallet == nil {
		if err := h.walletRepo.CreateWallet(ctx.Request.Context(), operation.WalletID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wallet"})
			return
		}
	}

	// Выполняем операцию
	if err := h.walletRepo.UpdateWalletBalance(ctx.Request.Context(), operation.WalletID, operation.OperationType, operation.Amount); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *WalletHandler) GetWalletBalance(ctx *gin.Context) {
	walletID := ctx.Param("walletId")

	if walletID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Wallet ID is required"})
		return
	}

	wallet, err := h.walletRepo.GetWallet(ctx.Request.Context(), walletID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get wallet"})
		return
	}

	if wallet == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
		return
	}

	ctx.JSON(http.StatusOK, models.WalletBalanceResponse{
		WalletID: wallet.ID,
		Balance:  wallet.Balance,
	})
}
