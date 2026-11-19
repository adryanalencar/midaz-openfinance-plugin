package providers

import (
	"testing"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/stretchr/testify/assert"
)

// TestNewProviderFactoryOpenFinanceSuccess tests creating direct Open Finance provider
func TestNewProviderFactoryOpenFinanceSuccess(t *testing.T) {
	cfg := &bootstrap.Config{
		Provider: "openfinance",
	}

	provider, err := NewProviderFactory(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "openfinance", provider.GetName())
}

// TestNewProviderFactoryPluggySuccess tests creating Pluggy aggregator provider
func TestNewProviderFactoryPluggySuccess(t *testing.T) {
	cfg := &bootstrap.Config{
		Provider: "pluggy",
	}

	provider, err := NewProviderFactory(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, "pluggy", provider.GetName())
}

// TestNewProviderFactoryUnsupportedError tests creating an unknown provider
func TestNewProviderFactoryUnsupportedError(t *testing.T) {
	cfg := &bootstrap.Config{
		Provider: "unknown_vendor",
	}

	provider, err := NewProviderFactory(cfg)

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "unsupported provider 'unknown_vendor'")
}
