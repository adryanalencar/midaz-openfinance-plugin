package workers

import (
	"context"
	"time"

	libLog "github.com/LerianStudio/lib-commons/v2/commons/log"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// SyncWorker continuously monitors active bank connections and imports transactions into Midaz.
type SyncWorker struct {
	cfg         *bootstrap.Config
	storageRepo domain.StorageRepository
	provider    domain.OpenFinanceProvider
	midazClient *midaz.Client
	logger      libLog.Logger
	triggerChan chan struct{}
}

// NewSyncWorker creates a new automated background synchronization worker.
func NewSyncWorker(
	cfg *bootstrap.Config,
	storageRepo domain.StorageRepository,
	provider domain.OpenFinanceProvider,
	midazClient *midaz.Client,
	logger libLog.Logger,
) *SyncWorker {
	return &SyncWorker{
		cfg:         cfg,
		storageRepo: storageRepo,
		provider:    provider,
		midazClient: midazClient,
		logger:      logger,
		triggerChan: make(chan struct{}, 10),
	}
}

// TriggerNow triggers an immediate sync cycle without waiting for the next timer tick.
func (w *SyncWorker) TriggerNow() {
	select {
	case w.triggerChan <- struct{}{}:
	default:
	}
}

// Start runs the background synchronization loop until context is canceled.
func (w *SyncWorker) Start(ctx context.Context) {
	interval := time.Duration(w.cfg.SyncIntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 60 * time.Minute
	}

	w.logger.Infof("Starting background Open Finance Sync Worker (Interval: %v)", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial immediate sync on worker startup
	w.runSyncCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			w.logger.Infof("Sync worker stopping due to context cancellation")
			return
		case <-ticker.C:
			w.runSyncCycle(ctx)
		case <-w.triggerChan:
			w.logger.Infof("Manual sync trigger received, executing sync cycle")
			w.runSyncCycle(ctx)
		}
	}
}

func (w *SyncWorker) runSyncCycle(ctx context.Context) {
	connections, err := w.storageRepo.ListActiveConnections(ctx)
	if err != nil {
		w.logger.Errorf("Failed to retrieve active bank connections: %v", err)
		return
	}

	if len(connections) == 0 {
		w.logger.Debugf("No active bank connections to sync")
		return
	}

	w.logger.Infof("Executing sync cycle across %d active bank connection(s)", len(connections))

	for _, conn := range connections {
		if ctx.Err() != nil {
			return
		}
		w.syncConnection(ctx, conn)
	}
}

func (w *SyncWorker) syncConnection(ctx context.Context, conn domain.BankConnection) {
	w.logger.Infof("Syncing connection ID %s (Bank COMPE: %s, Account: %s)", conn.ID, conn.BankCompe, conn.BankAccountID)

	checkpoint, err := w.storageRepo.GetCheckpoint(ctx, conn.ID)
	if err != nil {
		w.logger.Errorf("Error retrieving checkpoint for connection %s: %v", conn.ID, err)
		return
	}

	now := time.Now().UTC()
	var fromDate time.Time

	if checkpoint != nil && !checkpoint.LastBookingDate.IsZero() {
		// Lookback window: rewind by configured hours to catch delayed postings
		lookback := time.Duration(w.cfg.LookbackWindowHours) * time.Hour
		fromDate = checkpoint.LastBookingDate.Add(-lookback)
	} else {
		// Initial sync default: 90 days in the past
		fromDate = now.AddDate(0, -3, 0)
	}

	currentPage := 1
	pageSize := 100
	totalProcessed := 0
	latestBookingDate := fromDate

	for {
		page, err := w.provider.FetchTransactions(ctx, conn.BankAccountID, fromDate, now, currentPage, pageSize)
		if err != nil {
			w.logger.Errorf("Provider error fetching transactions for %s: %v", conn.BankAccountID, err)
			break
		}

		if page == nil || len(page.Transactions) == 0 {
			break
		}

		for _, tx := range page.Transactions {
			// 1. Check idempotency deduplication in local storage
			alreadyProcessed, checkErr := w.storageRepo.IsTransactionProcessed(ctx, tx.ID)
			if checkErr == nil && alreadyProcessed {
				continue // Skip previously recorded transaction
			}

			// 2. Post balanced double-entry transaction to Midaz
			postErr := w.midazClient.PostDoubleEntryTransaction(ctx, tx, conn.MidazAccountAlias)
			if postErr != nil {
				w.logger.Errorf("Failed to post transaction %s to Midaz: %v", tx.ID, postErr)
				continue
			}

			// 3. Mark transaction as processed
			_ = w.storageRepo.RecordProcessedTransaction(ctx, &domain.ProcessedTransaction{
				ConnectionID:         conn.ID,
				BankTransactionID:    tx.ID,
				MidazTransactionCode: "OFP-" + tx.ID,
				Amount:               tx.Amount,
				CreditDebit:          string(tx.CreditDebitType),
				SyncedAt:             now,
			})

			totalProcessed++
			if tx.BookingDateTime.After(latestBookingDate) {
				latestBookingDate = tx.BookingDateTime
			}
		}

		if page.TotalPages > 0 && currentPage >= page.TotalPages {
			break
		}
		currentPage++
	}

	// Update checkpoint
	_ = w.storageRepo.SaveCheckpoint(ctx, &domain.SyncCheckpoint{
		ConnectionID:    conn.ID,
		LastBookingDate: latestBookingDate,
		Status:          "SUCCESS",
		RecordsSynced:   totalProcessed,
		LastSyncedAt:    now,
	})

	w.logger.Infof("Finished sync for connection %s: %d new transaction(s) posted to Midaz", conn.ID, totalProcessed)
}
