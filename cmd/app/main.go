package main

import (
	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	httpIn "github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/in"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
)

func main() {
	// 1. Initialize local .env if present
	libCommons.InitLocalEnvConfig()

	// 2. Load configuration from environment variables
	cfg := &bootstrap.Config{}
	if err := libCommons.SetConfigFromEnvVars(cfg); err != nil {
		panic(err)
	}

	// 3. Structured Logging (Zap)
	logger := libZap.InitializeLogger()

	// 4. Initialize Provider (openfinance direct or pluggy aggregator)
	provider, err := providers.NewProviderFactory(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize provider: %v", err)
	}
	logger.Infof("Initialized banking provider: %s", provider.GetName())

	// 5. Initialize Midaz API Client
	_ = midaz.NewMidazClient(cfg)

	// 6. Setup HTTP In Adapters (Manifest + Sync + Health)
	manifestHandler := httpIn.NewManifestHandler(cfg)
	syncHandler := httpIn.NewSyncHandler()
	router := httpIn.NewRouter(manifestHandler, syncHandler)

	// 7. Start Service
	service := &bootstrap.Service{
		FiberApp: router,
		Config:   cfg,
		Logger:   logger,
	}

	service.Run()
}
