package command

import (
	"context"
	"fmt"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/google/uuid"
)

// CreatePluggyConnectionCommand contains inputs to register a newly connected Pluggy account.
type CreatePluggyConnectionCommand struct {
	ItemID                           string `json:"itemId"`
	AccountID                        string `json:"accountId"`
	MidazAccountAlias                string `json:"midazAccountAlias"`
	ClearingAccountAlias             string `json:"clearingAccountAlias,omitempty"`
	MotherAccountID                  string `json:"motherAccountId,omitempty"`
	MotherAccountAlias               string `json:"motherAccountAlias,omitempty"`
	ReconciliationEnabled            bool   `json:"reconciliationEnabled,omitempty"`
	ReconciliationReceivablesAccount string `json:"reconciliationReceivablesAccount,omitempty"`
	MatcherContextID                 string `json:"matcherContextId,omitempty"`
}

// CreatePluggyConnectionUseCase binds a Pluggy Item/Account to a Midaz Account Alias.
type CreatePluggyConnectionUseCase struct {
	pluggyClient *pluggy.PluggyClient
	storageRepo  domain.StorageRepository
}

// NewCreatePluggyConnectionUseCase creates a new connection registration command.
func NewCreatePluggyConnectionUseCase(
	pluggyClient *pluggy.PluggyClient,
	storageRepo domain.StorageRepository,
) *CreatePluggyConnectionUseCase {
	return &CreatePluggyConnectionUseCase{
		pluggyClient: pluggyClient,
		storageRepo:  storageRepo,
	}
}

// Execute validates and saves a new Pluggy bank connection.
func (uc *CreatePluggyConnectionUseCase) Execute(ctx context.Context, cmd *CreatePluggyConnectionCommand) (*domain.BankConnection, error) {
	if cmd.ItemID == "" || cmd.AccountID == "" || cmd.MidazAccountAlias == "" {
		return nil, fmt.Errorf("itemId, accountId, and midazAccountAlias are required")
	}

	// 1. Verify Item in Pluggy
	item, err := uc.pluggyClient.GetItem(ctx, cmd.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve item from Pluggy: %w", err)
	}

	// 2. Lookup Account in Pluggy
	accounts, err := uc.pluggyClient.ListAccounts(ctx, cmd.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts for item: %w", err)
	}

	var matchedAccount *domain.BankAccount
	for _, acc := range accounts {
		if acc.ID == cmd.AccountID {
			matchedAccount = &acc
			break
		}
	}
	if matchedAccount == nil && len(accounts) > 0 {
		matchedAccount = &accounts[0] // Fallback to primary account
	}

	bankName := "Pluggy Institution"
	if matchedAccount != nil && matchedAccount.BrandName != "" {
		bankName = matchedAccount.BrandName
	}

	// 3. Create domain connection
	conn := &domain.BankConnection{
		ID:                               uuid.New().String(),
		Provider:                         "pluggy",
		BankCompe:                        bankName,
		BankAccountID:                    cmd.AccountID,
		MidazAccountAlias:                cmd.MidazAccountAlias,
		MotherAccountID:                  cmd.MotherAccountID,
		MotherAccountAlias:               cmd.MotherAccountAlias,
		ReconciliationEnabled:            cmd.ReconciliationEnabled,
		ReconciliationReceivablesAccount: cmd.ReconciliationReceivablesAccount,
		MatcherContextID:                 cmd.MatcherContextID,
		ClearingAccountAlias:             cmd.ClearingAccountAlias,
		ConsentID:                         cmd.ItemID,
		Status:                           string(item.Status),
		CreatedAt:                        time.Now().UTC(),
		UpdatedAt:                        time.Now().UTC(),
	}

	if uc.storageRepo != nil {
		if err := uc.storageRepo.SaveConnection(ctx, conn); err != nil {
			return nil, fmt.Errorf("failed to save bank connection: %w", err)
		}

		// Initialize initial sync checkpoint
		_ = uc.storageRepo.SaveCheckpoint(ctx, &domain.SyncCheckpoint{
			ConnectionID: conn.ID,
			Status:       "IDLE",
			LastSyncedAt: time.Now().UTC(),
		})
	}

	return conn, nil
}
