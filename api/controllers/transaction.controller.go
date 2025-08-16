package controllers

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	transactionUseCase domain.TransactionUseCase
	ratingUseCase      domain.RatingUseCase
}

func NewTransactionController(tuc domain.TransactionUseCase, ruc domain.RatingUseCase) *TransactionController {
	return &TransactionController{
		transactionUseCase: tuc,
		ratingUseCase:      ruc,
	}
}

func (tc *TransactionController) AddTransaction(c *gin.Context) {
	var txn domain.Transaction
	if err := c.ShouldBindJSON(&txn); err != nil {
		c.JSON(config.GetStatusCode(config.ErrBadRequest), gin.H{"error": err.Error()})
		return
	}

	createdTxn, err := tc.transactionUseCase.AddTransaction(c.Request.Context(), &txn)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction added", "data": createdTxn})
}

func (tc *TransactionController) GetAllTransactions(c *gin.Context) {
	txns, err := tc.transactionUseCase.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": txns})
}

func (tc *TransactionController) GetTransactionsForCustomer(c *gin.Context) {
	customerID := c.Param("customerID")
	txns, err := tc.transactionUseCase.GetTransactionsForCustomer(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": txns})
}

func (tc *TransactionController) ImportTransactions(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file from form"})
		return
	}
	defer file.Close()

	ctx := c.Request.Context()
	transactions, ratings, logs, err := tc.ratingUseCase.ProcessTransactionsAndRatings(ctx, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Transactions imported and ratings calculated",
		"transactions": transactions,
		"ratings":      ratings,
		"logs":         logs,
	})
}
