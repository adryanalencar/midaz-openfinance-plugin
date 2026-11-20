package in

import (
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

type InternalHandler struct {
	storageRepo domain.StorageRepository
}

func NewInternalHandler(storageRepo domain.StorageRepository) *InternalHandler {
	return &InternalHandler{
		storageRepo: storageRepo,
	}
}

type MarkProcessedRequest struct {
	BankTransactionID    string          `json:"bankTransactionId"`
	MidazTransactionCode string          `json:"midazTransactionCode"`
	ConnectionID         string          `json:"connectionId,omitempty"`
	Amount               decimal.Decimal `json:"amount,omitempty"`
	CreditDebit          string          `json:"creditDebit,omitempty"`
	Status               string          `json:"status"` // "SETTLED_VIA_RECEIVABLES"
}

// MarkProcessed records a bank statement transaction as already reconciled and settled by the receivables plugin.
func (h *InternalHandler) MarkProcessed(c *fiber.Ctx) error {
	var req MarkProcessedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.BankTransactionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bankTransactionId is required",
		})
	}

	if h.storageRepo == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "acknowledged",
		})
	}

	// Idempotently record into processed_transactions
	_ = h.storageRepo.RecordProcessedTransaction(c.Context(), &domain.ProcessedTransaction{
		ConnectionID:         req.ConnectionID,
		BankTransactionID:    req.BankTransactionID,
		MidazTransactionCode: req.MidazTransactionCode,
		Amount:               req.Amount,
		CreditDebit:          req.CreditDebit,
		Status:               req.Status,
		SyncedAt:             time.Now().UTC(),
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":            "recorded",
		"bankTransactionId": req.BankTransactionID,
	})
}
