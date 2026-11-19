package in

import (
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/gofiber/fiber/v2"
)

// PluggyHandler exposes HTTP endpoints for managing Pluggy Connect Widget and account links.
type PluggyHandler struct {
	pluggyClient        *pluggy.PluggyClient
	createConnectionCmd *command.CreatePluggyConnectionUseCase
}

// NewPluggyHandler creates a new Pluggy HTTP handler.
func NewPluggyHandler(
	pluggyClient *pluggy.PluggyClient,
	createConnectionCmd *command.CreatePluggyConnectionUseCase,
) *PluggyHandler {
	return &PluggyHandler{
		pluggyClient:        pluggyClient,
		createConnectionCmd: createConnectionCmd,
	}
}

// CreateConnectToken generates a widget initialization token for Pluggy Connect.
// @Summary Generate Connect Token
// @Description Generates a temporary accessToken to initialize the client-side Pluggy Connect Widget
// @Tags Pluggy
// @Accept json
// @Produce json
// @Success 200 {object} pluggy.ConnectTokenResponse
// @Router /v1/connect-token/pluggy [post]
func (h *PluggyHandler) CreateConnectToken(c *fiber.Ctx) error {
	if h.pluggyClient == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "pluggy provider client is not configured",
		})
	}

	var req pluggy.ConnectTokenRequest
	_ = c.BodyParser(&req) // Optional request payload

	resp, err := h.pluggyClient.CreateConnectToken(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// CreateConnection registers a new bank connection from the Pluggy Connect Widget.
// @Summary Register Pluggy Connection
// @Description Binds a newly authorized Pluggy item and account to a Midaz Account Alias
// @Tags Pluggy
// @Accept json
// @Produce json
// @Success 201 {object} domain.BankConnection
// @Router /v1/connections/pluggy [post]
func (h *PluggyHandler) CreateConnection(c *fiber.Ctx) error {
	if h.createConnectionCmd == nil {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "pluggy connection use case is not configured",
		})
	}

	var cmd command.CreatePluggyConnectionCommand
	if err := c.BodyParser(&cmd); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid connection payload",
		})
	}

	conn, err := h.createConnectionCmd.Execute(c.Context(), &cmd)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conn)
}
