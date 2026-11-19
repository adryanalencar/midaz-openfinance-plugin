package command

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/pluggy"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreatePluggyConnection_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey": "pluggy-key"}`))

		case "/items/item-pluggy-01":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": "item-pluggy-01",
				"status": "UPDATED",
				"connectorId": 201
			}`))

		case "/accounts":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"results": [
					{
						"id": "acc-pluggy-01",
						"name": "Conta Corrente",
						"connector": {"name": "Banco do Brasil"}
					}
				]
			}`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	cfg := &bootstrap.Config{
		PluggyBaseURL:      mockServer.URL,
		PluggyClientID:     "client-123",
		PluggyClientSecret: "secret-123",
	}

	pluggyClient, err := pluggy.NewPluggyClient(cfg)
	require.NoError(t, err)

	storage := newMockWebhookStorage()
	uc := NewCreatePluggyConnectionUseCase(pluggyClient, storage)

	conn, err := uc.Execute(context.Background(), &CreatePluggyConnectionCommand{
		ItemID:            "item-pluggy-01",
		AccountID:         "acc-pluggy-01",
		MidazAccountAlias: "bb_operational_acc",
	})

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, "pluggy", conn.Provider)
	assert.Equal(t, "Banco do Brasil", conn.BankCompe)
	assert.Equal(t, "acc-pluggy-01", conn.BankAccountID)
	assert.Equal(t, "bb_operational_acc", conn.MidazAccountAlias)
	assert.Equal(t, "UPDATED", conn.Status)

	savedConn, err := storage.FindConnectionByID(context.Background(), conn.ID)
	assert.NoError(t, err)
	assert.NotNil(t, savedConn)
}
