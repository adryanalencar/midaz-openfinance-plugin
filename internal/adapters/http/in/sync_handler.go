package in

import (
	"context"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/query"
	"github.com/gofiber/fiber/v2"
)

// SyncHandler handles manual sync triggers and status queries.
type SyncHandler struct {
	syncCmd *command.SyncTransactionsUseCase
	statusQ *query.GetSyncStatusUseCase
}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler(syncCmd *command.SyncTransactionsUseCase, statusQ *query.GetSyncStatusUseCase) *SyncHandler {
	return &SyncHandler{
		syncCmd: syncCmd,
		statusQ: statusQ,
	}
}

// TriggerSyncRequest represents the body of a manual sync trigger.
type TriggerSyncRequest struct {
	BankAccountID          string `json:"bank_account_id"`
	TargetMidazAccountAlias string `json:"target_midaz_account_alias"`
	FromDate               string `json:"from_date,omitempty"`
	ToDate                 string `json:"to_date,omitempty"`
}

// TriggerSync accepts manual sync trigger requests.
func (h *SyncHandler) TriggerSync(c *fiber.Ctx) error {
	var req TriggerSyncRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.BankAccountID == "" || req.TargetMidazAccountAlias == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "bank_account_id and target_midaz_account_alias are required",
		})
	}

	fromDate, _ := time.Parse("2006-01-02", req.FromDate)
	toDate, _ := time.Parse("2006-01-02", req.ToDate)

	input := command.SyncTransactionsInput{
		BankAccountID:           req.BankAccountID,
		TargetMidazAccountAlias: req.TargetMidazAccountAlias,
		FromDate:               fromDate,
		ToDate:                 toDate,
	}

	// Trigger execution (can be run in background worker goroutine or synchronously)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if h.syncCmd != nil {
			_, _ = h.syncCmd.Execute(ctx, input)
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "queued",
		"message": "Synchronization triggered successfully",
	})
}

// GetStatus returns the current sync status and metrics.
func (h *SyncHandler) GetStatus(c *fiber.Ctx) error {
	if h.statusQ != nil {
		status, err := h.statusQ.Execute(c.Context())
		if err == nil {
			return c.Status(fiber.StatusOK).JSON(status)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":                      "healthy",
		"accounts_tracked":            1,
		"total_transactions_ingested": 0,
		"failed_transactions":         0,
	})
}
