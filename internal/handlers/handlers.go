package handlers

import (
	"fmt"
	"net/http"

	"github.com/NKV510/wallet-service/internal/models"
	"github.com/NKV510/wallet-service/internal/repository"
	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	repo *repository.WalletRepository
}

func NewWalletHandler(repo *repository.WalletRepository) *WalletHandler {
	return &WalletHandler{repo: repo}
}

func (h *WalletHandler) ProcessOperation(c *gin.Context) {
	var operation models.WalletOperation
	if err := c.BindJSON(&operation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if operation.WalletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet ID is required"})
		return
	}

	if operation.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	if operation.OperationType != models.DEPOSIT && operation.OperationType != models.WITHDRAW {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation type"})
		return
	}

	existingWallet, err := h.repo.GetWallet(c.Request.Context(), operation.WalletID)
	if err != nil && err.Error() != "wallet not found" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get wallet: %v", err)})
		return
	}

	if existingWallet == nil {
		if err := h.repo.CreateWallet(c.Request.Context(), operation.WalletID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create wallet: %v", err)})
			return
		}
	}

	if err := h.repo.UpdateWalletBalance(c.Request.Context(), operation.WalletID, operation.OperationType, operation.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *WalletHandler) GetWalletBalance(c *gin.Context) {
	walletID := c.Param("walletId")
	if walletID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Wallet ID is required"})
		return
	}

	wallet, err := h.repo.GetWallet(c.Request.Context(), walletID)
	if err != nil {
		if err.Error() == "wallet not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get wallet: %v", err)})
		return
	}

	response := models.WalletBalanceResponse{
		WalletID: wallet.ID,
		Balance:  wallet.Balance,
	}

	c.JSON(http.StatusOK, response)
}
