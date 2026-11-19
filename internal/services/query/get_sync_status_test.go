package query

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetSyncStatusExecuteSuccess tests querying the sync status snapshot
func TestGetSyncStatusExecuteSuccess(t *testing.T) {
	uc := NewGetSyncStatusUseCase("openfinance")
	status, err := uc.Execute(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "healthy", status.Status)
	assert.Equal(t, "openfinance", status.Provider)
}
