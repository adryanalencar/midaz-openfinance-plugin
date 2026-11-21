package in

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExternalHandler_NilClient(t *testing.T) {
	handler := NewExternalHandler(nil)
	app := fiber.New()
	app.Get("/v1/external-accounts", handler.ListExternalAccounts)

	req := httptest.NewRequest(http.MethodGet, "/v1/external-accounts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, float64(0), body["count"])
}

func TestRouter_ExternalAccounts(t *testing.T) {
	mh := NewManifestHandler(nil)
	eh := NewExternalHandler(nil)
	app := NewRouter(mh, nil, nil, nil, nil, eh)

	req := httptest.NewRequest(http.MethodGet, "/v1/external-accounts", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	reqAlias := httptest.NewRequest(http.MethodGet, "/v1/accounts/external", nil)
	respAlias, err := app.Test(reqAlias, -1)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, respAlias.StatusCode)
}

