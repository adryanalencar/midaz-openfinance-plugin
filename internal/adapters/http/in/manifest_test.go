package in

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/bootstrap"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// TestGetManifestSuccess is responsible to test the /api/manifest endpoint returns valid Midaz Console metadata
func TestGetManifestSuccess(t *testing.T) {
	cfg := &bootstrap.Config{
		PluginHost: "http://localhost:8085",
	}

	mh := NewManifestHandler(cfg)
	sh := NewSyncHandler(nil, nil)
	app := NewRouter(mh, sh, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/manifest", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var manifest map[string]any
	err = json.Unmarshal(body, &manifest)
	assert.NoError(t, err)

	assert.Equal(t, "plugin-openfinance", manifest["name"])
	assert.Equal(t, "Open Finance Sync", manifest["title"])
	assert.Equal(t, "/openfinance", manifest["route"])
	assert.Equal(t, "http://localhost:8085/remoteEntry.js", manifest["entry"])
	assert.Equal(t, "/health", manifest["healthcheck"])
	assert.Equal(t, "Landmark", manifest["icon"])
	assert.Equal(t, true, manifest["enabled"])
	assert.Equal(t, "Midaz Community", manifest["author"])
}

// TestHealthCheckSuccess is responsible to test the /health endpoint
func TestHealthCheckSuccess(t *testing.T) {
	cfg := &bootstrap.Config{PluginHost: "http://localhost:8085"}
	mh := NewManifestHandler(cfg)
	sh := NewSyncHandler(nil, nil)
	app := NewRouter(mh, sh, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req, -1)

	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
