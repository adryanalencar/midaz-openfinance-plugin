package in

import (
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/gofiber/fiber/v2"
)

// ManifestHandler handles dynamic discovery requests from Midaz Console.
type ManifestHandler struct {
	cfg *bootstrap.Config
}

// NewManifestHandler creates a new ManifestHandler.
func NewManifestHandler(cfg *bootstrap.Config) *ManifestHandler {
	return &ManifestHandler{cfg: cfg}
}

// GetManifest returns the PluginManifestEntity for the Midaz Console.
func (h *ManifestHandler) GetManifest(c *fiber.Ctx) error {
	manifest := fiber.Map{
		"name":        "plugin-openfinance",
		"title":       "Open Finance Sync",
		"description": "Automated bank statement mirroring and double-entry ledger reconciliation",
		"version":     "1.0.0",
		"route":       "/openfinance",
		"entry":       h.cfg.PluginHost + "/remoteEntry.js",
		"healthcheck": "/health",
		"host":        h.cfg.PluginHost,
		"icon":        "Landmark",
		"enabled":     true,
		"author":      "Midaz Community",
	}

	return c.Status(fiber.StatusOK).JSON(manifest)
}
