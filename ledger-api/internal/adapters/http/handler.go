package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/domain/transaction"
	"github.com/suphakin-th/ledger-service/ledger-api/internal/usecases"
)

type Handler struct {
	createAccount     *usecases.CreateAccountUseCase
	createTransaction *usecases.CreateTransactionUseCase
	getSummary        *usecases.GetSummaryUseCase
}

func NewHandler(
	createAccount *usecases.CreateAccountUseCase,
	createTransaction *usecases.CreateTransactionUseCase,
	getSummary *usecases.GetSummaryUseCase,
) *Handler {
	return &Handler{
		createAccount:     createAccount,
		createTransaction: createTransaction,
		getSummary:        getSummary,
	}
}

func (h *Handler) CreateAccount(c *gin.Context) {
	var body struct {
		Name     string `json:"name"     binding:"required"`
		Currency string `json:"currency" binding:"required,len=3"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	a, err := h.createAccount.Execute(c.Request.Context(), usecases.CreateAccountCommand{
		Name:     body.Name,
		Currency: body.Currency,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *Handler) CreateTransaction(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	var body struct {
		Type        string `json:"type"         binding:"required,oneof=credit debit"`
		AmountCents int64  `json:"amount_cents" binding:"required,gt=0"`
		Currency    string `json:"currency"     binding:"required,len=3"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.createTransaction.Execute(c.Request.Context(), usecases.CreateTransactionCommand{
		AccountID:   accountID,
		Type:        transaction.Type(body.Type),
		AmountCents: body.AmountCents,
		Currency:    body.Currency,
		Description: body.Description,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tx)
}

func (h *Handler) GetSummary(c *gin.Context) {
	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	summary, err := h.getSummary.Execute(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}
