package providers

import (
	"fmt"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// NewProviderFactory builds the designated provider based on the PROVIDER environment variable.
func NewProviderFactory(cfg *bootstrap.Config) (domain.OpenFinanceProvider, error) {
	switch cfg.Provider {
	case "openfinance":
		return openfinance.NewDirectOpenFinanceClient(cfg)
	case "pluggy":
		return pluggy.NewPluggyClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider '%s': must be 'openfinance' or 'pluggy'", cfg.Provider)
	}
}
