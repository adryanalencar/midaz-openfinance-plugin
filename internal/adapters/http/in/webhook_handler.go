package in

import (
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/gofiber/fiber/v2"
)

// WebhookHandler handles inbound webhook notifications from external providers like Pluggy.
type WebhookHandler struct {
	pluggyWebhookCmd *command.HandlePluggyWebhookUseCase
}

// NewWebhookHandler creates a new webhook HTTP handler.
func NewWebhookHandler(pluggyWebhookCmd *command.HandlePluggyWebhookUseCase) *WebhookHandler {
	return &WebhookHandler{
		pluggyWebhookCmd: pluggyWebhookCmd,
	}
}

// HandlePluggyWebhook receives and processes webhook events from Pluggy.
// @Summary Handle Pluggy Webhook
// @Description Ingests event-driven webhook notifications from the Pluggy aggregation platform
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /webhooks/pluggy [post]
func (h *WebhookHandler) HandlePluggyWebhook(c *fiber.Ctx) error {
	var event pluggy.WebhookEvent
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid webhook json payload",
		})
	}

	if event.Event == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing event type",
		})
	}

	if h.pluggyWebhookCmd != nil {
		if err := h.pluggyWebhookCmd.Execute(c.Context(), &event); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "received",
		"eventId": event.EventID,
	})
}
