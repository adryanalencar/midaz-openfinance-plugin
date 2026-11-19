package command

import (
	"context"
	"fmt"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// HandlePluggyWebhookUseCase processes event notifications from the Pluggy platform.
type HandlePluggyWebhookUseCase struct {
	cfg          *bootstrap.Config
	pluggyClient *pluggy.PluggyClient
	storageRepo  domain.StorageRepository
	midazClient  *midaz.Client
	logger       libLog.Logger
}

// NewHandlePluggyWebhookUseCase creates a new webhook processing command.
func NewHandlePluggyWebhookUseCase(
	cfg *bootstrap.Config,
	pluggyClient *pluggy.PluggyClient,
	storageRepo domain.StorageRepository,
	midazClient *midaz.Client,
	logger libLog.Logger,
) *HandlePluggyWebhookUseCase {
	if logger == nil {
		logger = &libLog.NoneLogger{}
	}
	return &HandlePluggyWebhookUseCase{
		cfg:          cfg,
		pluggyClient: pluggyClient,
		storageRepo:  storageRepo,
		midazClient:  midazClient,
		logger:       logger,
	}
}

// Execute processes a Pluggy webhook notification.
func (uc *HandlePluggyWebhookUseCase) Execute(ctx context.Context, event *pluggy.WebhookEvent) error {
	// 1. Security Check: Validate Client ID
	if uc.cfg.PluggyClientID != "" && event.ClientID != "" && event.ClientID != uc.cfg.PluggyClientID {
		return fmt.Errorf("unauthorized webhook client_id: %s", event.ClientID)
	}

	uc.logger.Infof("Processing Pluggy webhook event: %s (eventId: %s, itemId: %s)", event.Event, event.EventID, event.ItemID)

	switch event.Event {
	case "transactions/created":
		return uc.handleTransactionsCreated(ctx, event)

	case "item/updated", "item/login_succeeded":
		return uc.updateItemStatus(ctx, event.ItemID, "ACTIVE")

	case "item/error":
		return uc.updateItemStatus(ctx, event.ItemID, "LOGIN_ERROR")

	case "item/waiting_user_input", "item/waiting_user_action":
		return uc.updateItemStatus(ctx, event.ItemID, "WAITING_USER_INPUT")

	case "item/deleted":
		return uc.updateItemStatus(ctx, event.ItemID, "DISCONNECTED")

	default:
		uc.logger.Infof("Ignoring unhandled Pluggy event: %s", event.Event)
		return nil
	}
}

func (uc *HandlePluggyWebhookUseCase) handleTransactionsCreated(ctx context.Context, event *pluggy.WebhookEvent) error {
	if uc.storageRepo == nil {
		uc.logger.Warn("Storage repository unavailable; skipping webhook transaction persistence")
		return nil
	}

	// 1. Find the local bank connection for this account/item
	conn, err := uc.storageRepo.FindConnectionByAccountID(ctx, event.AccountID)
	if err != nil {
		return fmt.Errorf("failed to lookup bank connection for account %s: %w", event.AccountID, err)
	}
	if conn == nil {
		// Try by Item ID
		conn, err = uc.storageRepo.FindConnectionByID(ctx, event.ItemID)
		if err != nil || conn == nil {
			uc.logger.Warnf("No active bank connection found for webhook accountId %s (itemId: %s)", event.AccountID, event.ItemID)
			return nil
		}
	}

	// 2. Fetch new transactions using the cursor link provided by Pluggy
	link := event.GetTransactionsLink()
	var txs []domain.BankTransaction
	if link != "" {
		txs, err = uc.pluggyClient.FetchAllTransactionsFromLink(ctx, link)
		if err != nil {
			return fmt.Errorf("failed to fetch transactions from link %s: %w", link, err)
		}
	} else if event.AccountID != "" {
		page, err := uc.pluggyClient.FetchTransactions(ctx, event.AccountID, time.Time{}, time.Time{}, 1, 100)
		if err != nil {
			return fmt.Errorf("failed to fetch transactions for account %s: %w", event.AccountID, err)
		}
		txs = page.Transactions
	}

	uc.logger.Infof("Fetched %d transactions from Pluggy webhook for account %s", len(txs), event.AccountID)

	// 3. Ingest each transaction into Midaz Ledger with double-entry balancing
	syncedCount := 0
	for _, tx := range txs {
		processed, err := uc.storageRepo.IsTransactionProcessed(ctx, tx.ID)
		if err != nil {
			uc.logger.Errorf("Failed to check transaction deduplication for %s: %v", tx.ID, err)
			continue
		}
		if processed {
			continue
		}

		midazTxCode := fmt.Sprintf("OFP-PLG-%s", tx.ID)
		if err := uc.midazClient.PostDoubleEntryTransaction(ctx, tx, conn.MidazAccountAlias); err != nil {
			uc.logger.Errorf("Failed to post double-entry transaction %s to Midaz: %v", tx.ID, err)
			continue
		}

		_ = uc.storageRepo.RecordProcessedTransaction(ctx, &domain.ProcessedTransaction{
			ConnectionID:         conn.ID,
			BankTransactionID:    tx.ID,
			MidazTransactionCode: midazTxCode,
			Amount:               tx.Amount,
			CreditDebit:          string(tx.CreditDebitType),
			SyncedAt:             time.Now().UTC(),
		})

		syncedCount++
	}

	uc.logger.Infof("Successfully mirrored %d new transactions to Midaz from webhook", syncedCount)

	// 4. Update sync checkpoint
	_ = uc.storageRepo.SaveCheckpoint(ctx, &domain.SyncCheckpoint{
		ConnectionID:  conn.ID,
		Status:        "SUCCESS",
		RecordsSynced: syncedCount,
		LastSyncedAt:  time.Now().UTC(),
	})

	return nil
}

func (uc *HandlePluggyWebhookUseCase) updateItemStatus(ctx context.Context, itemID, status string) error {
	if uc.storageRepo == nil {
		return nil
	}
	uc.logger.Infof("Updating Pluggy item %s status to %s", itemID, status)
	return uc.storageRepo.UpdateConnectionStatus(ctx, itemID, status)
}
