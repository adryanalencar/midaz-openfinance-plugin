package main

import (
	"context"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	httpIn "github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/in"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/postgres"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/command"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/query"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/workers"
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

	// 4. Initialize PostgreSQL Database & Storage Repository
	var storageRepo domain.StorageRepository
	db, err := postgres.ConnectPostgres(cfg.PostgreSQLDSN())
	if err != nil {
		logger.Warnf("PostgreSQL database connection deferred: %v", err)
	} else {
		logger.Infof("Connected to PostgreSQL database at %s:%s", cfg.DBHost, cfg.DBPort)
		storageRepo = postgres.NewPostgresRepository(db)
	}

	// 5. Initialize Provider (openfinance direct or pluggy aggregator)
	provider, err := providers.NewProviderFactory(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize provider: %v", err)
	}
	logger.Infof("Initialized banking provider: %s", provider.GetName())

	// 6. Initialize Midaz API Client
	midazClient := midaz.NewMidazClient(cfg)

	// 7. Bootstrap: Ensure dedicated clearing account exists (@external/BRL:openfinance)
	ensureClearingAccount := command.NewEnsureExternalAccountUseCase(midazClient)
	if err := ensureClearingAccount.Execute(context.Background()); err != nil {
		logger.Warnf("Clearing account verification deferred: %v", err)
	} else {
		logger.Infof("Verified @external clearing account: %s", cfg.MidazClearingAccountAlias)
	}

	// 8. Initialize CQRS Application Services
	syncTransactionsCmd := command.NewSyncTransactionsUseCase(provider, midazClient)
	getSyncStatusQuery := query.NewGetSyncStatusUseCase(provider.GetName())

	// 9. Setup Background Autonomous Sync Worker
	var syncWorker *workers.SyncWorker
	if storageRepo != nil {
		syncWorker = workers.NewSyncWorker(cfg, storageRepo, provider, midazClient, logger)
	}

	// 10. Setup HTTP In Adapters (Manifest + Sync + Health)
	manifestHandler := httpIn.NewManifestHandler(cfg)
	syncHandler := httpIn.NewSyncHandler(syncTransactionsCmd, getSyncStatusQuery)
	router := httpIn.NewRouter(manifestHandler, syncHandler)

	// 11. Start Service with graceful shutdown
	service := &bootstrap.Service{
		FiberApp: router,
		Config:   cfg,
		Logger:   logger,
		Worker:   syncWorker,
	}

	service.Run()
}
