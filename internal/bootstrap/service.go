package bootstrap

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/gofiber/fiber/v2"
)

// Service encapsulates the lifecycle of the Open Finance Plugin.
type Service struct {
	FiberApp *fiber.App
	Config   *Config
	Logger   libLog.Logger
}

// Run starts the Fiber HTTP server and handles graceful shutdown.
func (s *Service) Run() {
	s.Logger.Infof("Starting %s on %s (Provider: %s)", ApplicationName, s.Config.ServerAddress, s.Config.Provider)

	// Graceful shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.FiberApp.Listen(s.Config.ServerAddress); err != nil {
			s.Logger.Errorf("Server error: %v", err)
		}
	}()

	<-shutdown
	s.Logger.Infof("Shutting down %s gracefully...", ApplicationName)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.FiberApp.ShutdownWithContext(ctx); err != nil {
		s.Logger.Errorf("Error during shutdown: %v", err)
	}

	fmt.Println("Server stopped.")
}
