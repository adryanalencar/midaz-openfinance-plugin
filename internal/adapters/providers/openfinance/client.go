package openfinance

import (
	"context"
	"net/http"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/crypto"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/statement"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/domain"
)

// DirectOpenFinanceClient implements domain.OpenFinanceProvider for direct Open Finance Brasil APIs.
type DirectOpenFinanceClient struct {
	cfg             *bootstrap.Config
	statementClient *statement.Client
}

// NewDirectOpenFinanceClient initializes the mTLS and JWS client for Open Finance Brasil.
func NewDirectOpenFinanceClient(cfg *bootstrap.Config) (*DirectOpenFinanceClient, error) {
	var httpClient *http.Client
	var err error

	// If mTLS certificates are supplied, construct the secure mTLS client
	if cfg.OpenFinanceTransportCert != "" && cfg.OpenFinanceTransportKey != "" {
		httpClient, err = crypto.BuildMTLSClient(crypto.TLSConfigOptions{
			CertPath: cfg.OpenFinanceTransportCert,
			KeyPath:  cfg.OpenFinanceTransportKey,
			Timeout:  30 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Fallback standard HTTP client (e.g. for testing / local sandbox mocks)
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	statementClient := statement.NewStatementClient(httpClient, cfg.OpenFinanceDirectoryURL)

	return &DirectOpenFinanceClient{
		cfg:             cfg,
		statementClient: statementClient,
	}, nil
}

func (c *DirectOpenFinanceClient) GetName() string {
	return "openfinance"
}

func (c *DirectOpenFinanceClient) ListAccounts(ctx context.Context, connectionID string) ([]domain.BankAccount, error) {
	return c.statementClient.ListAccounts(ctx, connectionID)
}

func (c *DirectOpenFinanceClient) FetchTransactions(ctx context.Context, accountID string, fromDate, toDate time.Time, page, pageSize int) (*domain.TransactionPage, error) {
	return c.statementClient.FetchTransactions(ctx, "", accountID, fromDate, toDate, page, pageSize, "")
}

func (c *DirectOpenFinanceClient) HealthCheck(ctx context.Context) error {
	return nil
}
