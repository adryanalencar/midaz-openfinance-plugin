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

// BackgroundWorker represents a runnable background process.
type BackgroundWorker interface {
	Start(ctx context.Context)
}

// Service encapsulates the lifecycle of the Open Finance Plugin.
type Service struct {
	FiberApp *fiber.App
	Config   *Config
	Logger   libLog.Logger
	Worker   BackgroundWorker
}

// Run starts the Fiber HTTP server, background worker, and handles graceful shutdown.
func (s *Service) Run() {
	s.Logger.Infof("Starting %s on %s (Provider: %s)", ApplicationName, s.Config.ServerAddress, s.Config.Provider)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Start Background Sync Worker
	if s.Worker != nil {
		go s.Worker.Start(ctx)
	}

	// 2. Graceful shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := s.FiberApp.Listen(s.Config.ServerAddress); err != nil {
			s.Logger.Errorf("Server error: %v", err)
		}
	}()

	<-shutdown
	s.Logger.Infof("Shutting down %s gracefully...", ApplicationName)

	// Stop background workers
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := s.FiberApp.ShutdownWithContext(shutdownCtx); err != nil {
		s.Logger.Errorf("Error during shutdown: %v", err)
	}

	fmt.Println("Server stopped cleanly.")
}
