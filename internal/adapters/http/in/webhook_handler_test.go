package in

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHandlePluggyWebhookEndpoint(t *testing.T) {
	cfg := &bootstrap.Config{PluginHost: "http://localhost:8085"}
	mh := NewManifestHandler(cfg)
	sh := NewSyncHandler(nil, nil)
	wh := NewWebhookHandler(nil)
	app := NewRouter(mh, sh, wh, nil)

	// 1. Invalid JSON
	reqBad := httptest.NewRequest(http.MethodPost, "/webhooks/pluggy", bytes.NewBuffer([]byte(`invalid-json`)))
	reqBad.Header.Set("Content-Type", "application/json")
	respBad, err := app.Test(reqBad, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, respBad.StatusCode)

	// 2. Missing event type
	reqEmpty := httptest.NewRequest(http.MethodPost, "/webhooks/pluggy", bytes.NewBuffer([]byte(`{}`)))
	reqEmpty.Header.Set("Content-Type", "application/json")
	respEmpty, err := app.Test(reqEmpty, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, respEmpty.StatusCode)

	// 3. Valid webhook
	validJSON := `{"event": "transactions/created", "eventId": "evt-001", "itemId": "item-001"}`
	reqValid := httptest.NewRequest(http.MethodPost, "/webhooks/pluggy", bytes.NewBuffer([]byte(validJSON)))
	reqValid.Header.Set("Content-Type", "application/json")
	respValid, err := app.Test(reqValid, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, respValid.StatusCode)
}
