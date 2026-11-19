package in

import (
	"github.com/gofiber/fiber/v2"
)

// SyncHandler handles manual sync triggers and status queries.
type SyncHandler struct{}

// NewSyncHandler creates a new SyncHandler.
func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

// TriggerSync accepts manual sync trigger requests.
func (h *SyncHandler) TriggerSync(c *fiber.Ctx) error {
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "queued",
		"message": "Synchronization triggered successfully",
	})
}

// GetStatus returns the current sync status and metrics.
func (h *SyncHandler) GetStatus(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":                      "healthy",
		"accounts_tracked":            1,
		"total_transactions_ingested": 0,
		"failed_transactions":         0,
	})
}
