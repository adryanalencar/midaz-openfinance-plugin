package command

import (
	"context"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/http/out/midaz"
)

// EnsureExternalAccountUseCase guarantees the @external clearing account is available in Midaz.
type EnsureExternalAccountUseCase struct {
	midazClient *midaz.Client
}

// NewEnsureExternalAccountUseCase creates a new EnsureExternalAccountUseCase.
func NewEnsureExternalAccountUseCase(midazClient *midaz.Client) *EnsureExternalAccountUseCase {
	return &EnsureExternalAccountUseCase{
		midazClient: midazClient,
	}
}

// Execute checks and provisions the external clearing account.
func (uc *EnsureExternalAccountUseCase) Execute(ctx context.Context) error {
	return uc.midazClient.EnsureClearingAccount(ctx)
}
