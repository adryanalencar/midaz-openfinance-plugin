# Multi-Provider Strategy & Extensibility

## 1. Design Motivation

Not all deployment environments possess active BACEN participant status or production ICP-Brasil certificates on day one. Conversely, organizations operating at high volume need direct bank connectivity to eliminate third-party per-transaction SaaS fees.

To solve both realities seamlessly, the plugin abstracts external connectivity behind the **`OpenFinanceProvider`** interface. The actual provider is selected dynamically at runtime via the `PROVIDER` environment variable:

```
PROVIDER=openfinance  # Direct Open Finance Brasil (mTLS + FAPI)
PROVIDER=pluggy       # Pluggy Aggregator API
```

---

## 2. Comparison Matrix: Direct vs. Aggregator

| Feature | Direct Open Finance (`openfinance`) | Aggregator (`pluggy`) |
| :--- | :--- | :--- |
| **Prerequisites** | Regulated BACEN participant status + ICP-Brasil certificates | Standard SaaS API Key (`CLIENT_ID` + `CLIENT_SECRET`) |
| **Cost Profile** | Zero per-transaction fees; infrastructure costs only | Per-connection / per-sync API fees |
| **Setup Time** | Weeks (Directory onboarding, DCR testing) | Minutes |
| **Bank Coverage** | Regulated institutions in Brazil | Brazilian banks, fintechs, and foreign accounts |
| **Maintenance** | In-house certificate lifecycle and bank API versioning | Managed by aggregator vendor |

---

## 3. The `OpenFinanceProvider` Domain Interface

Both direct and aggregator drivers must satisfy the domain contract defined in `internal/domain/provider.go`:

```go
package domain

import (
	"context"
	"time"
)

// OpenFinanceProvider defines the uniform contract for banking integrations.
type OpenFinanceProvider interface {
	// GetName returns the provider identifier (e.g., "openfinance", "pluggy").
	GetName() string

	// ListAccounts returns all active bank accounts under the current authorization.
	ListAccounts(ctx context.Context, connectionID string) ([]BankAccount, error)

	// FetchTransactions retrieves historical transactions within a date window.
	FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*TransactionPage, error)

	// HealthCheck validates connectivity to the upstream provider.
	HealthCheck(ctx context.Context) error
}
```

---

## 4. The Provider Factory

The application service uses a factory pattern to instantiate the appropriate adapter during startup:

```go
package providers

import (
	"fmt"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
)

func NewProviderFactory(cfg *bootstrap.Config) (domain.OpenFinanceProvider, error) {
	switch cfg.Provider {
	case "openfinance":
		return openfinance.NewDirectOpenFinanceClient(cfg)
	case "pluggy":
		return pluggy.NewPluggyClient(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s (supported: 'openfinance', 'pluggy')", cfg.Provider)
	}
}
```

---

## 5. Adding a New Provider (e.g., Belvo or Celcoin)

To add another aggregation or banking provider:
1. Create a new subpackage under `internal/adapters/providers/<provider_name>`.
2. Implement the `domain.OpenFinanceProvider` interface methods.
3. Map upstream vendor JSON payloads into `domain.BankTransaction`.
4. Register the new provider key in `NewProviderFactory`.
5. Add vendor-specific environment variables to `internal/bootstrap/config.go` with `default:""`.
