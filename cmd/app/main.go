package main

import (
	"context"

	libCommons "github.com/LerianStudio/lib-commons/v2/commons"
	libZap "github.com/LerianStudio/lib-commons/v2/commons/zap"
	httpIn "github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/in"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/postgres"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
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

	// 4. Initialize PostgreSQL Database, Run Migrations & Setup Repository
	var storageRepo domain.StorageRepository
	db, err := postgres.ConnectPostgres(cfg.PostgreSQLDSN())
	if err != nil {
		logger.Warnf("PostgreSQL database connection deferred: %v", err)
	} else {
		if cfg.IsExternalDatabase() {
			logger.Infof("Connected to external PostgreSQL database via URL connection string")
		} else {
			logger.Infof("Connected to PostgreSQL database at %s:%s (database: %s)", cfg.DBHost, cfg.DBPort, cfg.DBName)
		}

		// Database Migration Governance: auto-migrate on startup if enabled
		if cfg.DBAutoMigrate {
			logger.Infof("Executing database migration governance check...")
			migrator := postgres.NewMigrator(db, logger)
			if err := migrator.MigrateUp(context.Background()); err != nil {
				logger.Fatalf("Database migration failed: %v", err)
			}
		} else {
			logger.Infof("Database auto-migration is disabled (DB_AUTO_MIGRATE=false)")
		}

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

	// 10. Setup Pluggy & Webhook Handlers
	var webhookHandler *httpIn.WebhookHandler
	var pluggyHandler *httpIn.PluggyHandler
	if cfg.PluggyClientID != "" || cfg.Provider == "pluggy" {
		pluggyClient, err := pluggy.NewPluggyClient(cfg)
		if err == nil {
			handlePluggyWebhookCmd := command.NewHandlePluggyWebhookUseCase(cfg, pluggyClient, storageRepo, midazClient, logger)
			createPluggyConnCmd := command.NewCreatePluggyConnectionUseCase(pluggyClient, storageRepo)
			webhookHandler = httpIn.NewWebhookHandler(handlePluggyWebhookCmd)
			pluggyHandler = httpIn.NewPluggyHandler(pluggyClient, createPluggyConnCmd)
			logger.Infof("Initialized Pluggy Connect & Webhook handlers")
		}
	}

	// 11. Setup HTTP In Adapters (Manifest + Sync + Webhooks + Pluggy + Internal + External + Health)
	manifestHandler := httpIn.NewManifestHandler(cfg)
	syncHandler := httpIn.NewSyncHandler(syncTransactionsCmd, getSyncStatusQuery)
	internalHandler := httpIn.NewInternalHandler(storageRepo)
	externalHandler := httpIn.NewExternalHandler(midazClient)
	router := httpIn.NewRouter(manifestHandler, syncHandler, webhookHandler, pluggyHandler, internalHandler, externalHandler)

	// 11. Start Service with graceful shutdown
	service := &bootstrap.Service{
		FiberApp: router,
		Config:   cfg,
		Logger:   logger,
		Worker:   syncWorker,
	}

	service.Run()
}
