package in

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// NewRouter registers all HTTP routes for the Open Finance Plugin.
func NewRouter(mh *ManifestHandler, sh *SyncHandler) *fiber.App {
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

	// Sync API
	v1 := app.Group("/v1")
	v1.Post("/sync", sh.TriggerSync)
	v1.Get("/sync/status", sh.GetStatus)

	return app
}
