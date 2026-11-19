package crypto

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestInjectFAPIHeadersSuccess tests that all mandatory FAPI headers are populated
func TestInjectFAPIHeadersSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://api.banco.com.br/accounts", nil)

	interactionID := InjectFAPIHeaders(req, "200.189.12.5", "CustomAgent/1.0")

	assert.NotEmpty(t, interactionID)
	_, err := uuid.Parse(interactionID)
	assert.NoError(t, err, "interactionID must be a valid UUID v4")

	assert.Equal(t, interactionID, req.Header.Get(HeaderXFapiInteractionID))
	assert.Equal(t, "200.189.12.5", req.Header.Get(HeaderXFapiCustomerIPAddress))
	assert.Equal(t, "CustomAgent/1.0", req.Header.Get(HeaderXCustomerUserAgent))
	assert.Equal(t, "application/json", req.Header.Get("Accept"))

	// Validate date format is RFC 7231 (HTTP-date)
	dateStr := req.Header.Get(HeaderXFapiAuthDate)
	assert.NotEmpty(t, dateStr)
	_, err = time.Parse(http.TimeFormat, dateStr)
	assert.NoError(t, err, "x-fapi-auth-date must adhere to RFC 7231")
}

// TestExtractInteractionID tests extracting the header from responses
func TestExtractInteractionID(t *testing.T) {
	resp := &http.Response{
		Header: make(http.Header),
	}
	expectedID := "01934371-3cb5-71bb-98f5-66778899aabb"
	resp.Header.Set(HeaderXFapiInteractionID, expectedID)

	extracted := ExtractInteractionID(resp)
	assert.Equal(t, expectedID, extracted)

	assert.Empty(t, ExtractInteractionID(nil))
}
