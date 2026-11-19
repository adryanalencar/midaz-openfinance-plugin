package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// NewRouter registers all HTTP routes for the Open Finance Plugin.
func NewRouter(mh *ManifestHandler, sh *SyncHandler, wh *WebhookHandler, ph *PluggyHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "healthy",
			"app":    "midaz-plugin-openfinance",
		})
	})

	// Midaz Console Manifest
	app.Get("/api/manifest", mh.GetManifest)

	// Webhook endpoints
	if wh != nil {
		app.Post("/webhooks/pluggy", wh.HandlePluggyWebhook)
	}

	// Core API Group
	v1 := app.Group("/v1")
	v1.Post("/sync", sh.TriggerSync)
	v1.Get("/sync/status", sh.GetStatus)

	// Pluggy Connect endpoints
	if ph != nil {
		v1.Post("/connect-token/pluggy", ph.CreateConnectToken)
		v1.Post("/connections/pluggy", ph.CreateConnection)
	}

	return app
}
