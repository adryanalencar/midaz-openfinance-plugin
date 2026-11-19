package main

import (
	"context"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	httpIn "github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/in"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/query"
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
	midazClient := midaz.NewMidazClient(cfg)

	// 6. Bootstrap: Ensure dedicated clearing account exists (@external/BRL:openfinance)
	ensureClearingAccount := command.NewEnsureExternalAccountUseCase(midazClient)
	if err := ensureClearingAccount.Execute(context.Background()); err != nil {
		logger.Warnf("Clearing account verification deferred: %v", err)
	} else {
		logger.Infof("Verified @external clearing account: %s", cfg.MidazClearingAccountAlias)
	}

	// 7. Initialize CQRS Application Services
	syncTransactionsCmd := command.NewSyncTransactionsUseCase(provider, midazClient)
	getSyncStatusQuery := query.NewGetSyncStatusUseCase(provider.GetName())

	// 8. Setup HTTP In Adapters (Manifest + Sync + Health)
	manifestHandler := httpIn.NewManifestHandler(cfg)
	syncHandler := httpIn.NewSyncHandler(syncTransactionsCmd, getSyncStatusQuery)
	router := httpIn.NewRouter(manifestHandler, syncHandler)

	// 9. Start Service with graceful shutdown
	service := &bootstrap.Service{
		FiberApp: router,
		Config:   cfg,
		Logger:   logger,
	}

	service.Run()
}
