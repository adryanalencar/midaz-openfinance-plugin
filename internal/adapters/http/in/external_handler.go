package in

import (
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/gofiber/fiber/v2"
)

// ExternalHandler handles HTTP requests relating to Midaz external clearing accounts.
type ExternalHandler struct {
	midazClient *midaz.Client
}

// NewExternalHandler creates a new ExternalHandler.
func NewExternalHandler(midazClient *midaz.Client) *ExternalHandler {
	return &ExternalHandler{
		midazClient: midazClient,
	}
}

// ListExternalAccounts retrieves all external clearing accounts configured in the Midaz ledger.
// @Summary List External Accounts
// @Description Returns all accounts of type 'external' in the target Midaz ledger
// @Tags Accounts
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/external-accounts [get]
func (h *ExternalHandler) ListExternalAccounts(c *fiber.Ctx) error {
	if h.midazClient == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"items": []any{},
			"count": 0,
		})
	}

	accounts, err := h.midazClient.ListExternalAccounts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"items": accounts,
		"count": len(accounts),
	})
}
