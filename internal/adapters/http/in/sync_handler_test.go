package in

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/LerianStudio/midaz-plugin-openfinance/internal/services/query"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestTriggerSyncValidation tests payload validation on POST /v1/sync
func TestTriggerSyncValidation(t *testing.T) {
	cfg := &bootstrap.Config{PluginHost: "http://localhost:8085"}
	mh := NewManifestHandler(cfg)
	statusQ := query.NewGetSyncStatusUseCase("openfinance")
	sh := NewSyncHandler(nil, statusQ)
	app := NewRouter(mh, sh, nil, nil, nil)

	// 1. Missing required fields
	req := httptest.NewRequest(http.MethodPost, "/v1/sync", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	// 2. Valid payload accepted
	validBody := map[string]string{
		"bank_account_id":            "acc-001",
		"target_midaz_account_alias": "midaz_itau_acc",
	}
	bodyBytes, _ := json.Marshal(validBody)

	reqValid := httptest.NewRequest(http.MethodPost, "/v1/sync", bytes.NewBuffer(bodyBytes))
	reqValid.Header.Set("Content-Type", "application/json")
	respValid, err := app.Test(reqValid, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusAccepted, respValid.StatusCode)
}

// TestGetSyncStatusEndpoint tests GET /v1/sync/status
func TestGetSyncStatusEndpoint(t *testing.T) {
	cfg := &bootstrap.Config{PluginHost: "http://localhost:8085"}
	mh := NewManifestHandler(cfg)
	statusQ := query.NewGetSyncStatusUseCase("openfinance")
	sh := NewSyncHandler(nil, statusQ)
	app := NewRouter(mh, sh, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/sync/status", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
