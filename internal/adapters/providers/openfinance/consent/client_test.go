package consent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSigner(t *testing.T) *crypto.JWSSigner {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	signer, err := crypto.NewJWSSigner("test-key-id", "test-client-id", keyPEM)
	require.NoError(t, err)
	return signer
}

// TestCreateConsentSuccess tests submitting a valid consent request
func TestCreateConsentSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/open-banking/consents/v3/consents", r.URL.Path)
		assert.NotEmpty(t, r.Header.Get(crypto.HeaderXFapiInteractionID))

		var req CreateConsentRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)

		assert.Contains(t, req.Data.Permissions, "ACCOUNTS_READ")
		assert.Contains(t, req.Data.Permissions, "ACCOUNTS_TRANSACTIONS_READ")

		resp := ConsentResponse{
			Data: ConsentResponseData{
				ConsentID:        "urn:banco:018f9211-consent",
				Status:           "AWAITING_AUTHORISATION",
				CreationDateTime: time.Now().UTC(),
			},
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	signer := setupTestSigner(t)
	client := NewConsentClient(mockServer.Client(), signer, mockServer.URL, mockServer.URL+"/token", "test-client")

	payload := BuildStandardConsentPayload(time.Now().Add(365 * 24 * time.Hour))
	res, err := client.CreateConsent(context.Background(), "mock-token", payload)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "urn:banco:018f9211-consent", res.Data.ConsentID)
	assert.Equal(t, "AWAITING_AUTHORISATION", res.Data.Status)
}

// TestGetConsentSuccess tests querying an existing consent status
func TestGetConsentSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/open-banking/consents/v3/consents/consent-xyz", r.URL.Path)

		resp := ConsentResponse{
			Data: ConsentResponseData{
				ConsentID: "consent-xyz",
				Status:    "AUTHORISED",
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	signer := setupTestSigner(t)
	client := NewConsentClient(mockServer.Client(), signer, mockServer.URL, mockServer.URL+"/token", "test-client")

	res, err := client.GetConsent(context.Background(), "mock-token", "consent-xyz")

	assert.NoError(t, err)
	assert.Equal(t, "AUTHORISED", res.Data.Status)
}

// TestExchangeCodeForTokenSuccess tests OAuth2 token exchange with private_key_jwt assertion
func TestExchangeCodeForTokenSuccess(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		err := r.ParseForm()
		assert.NoError(t, err)

		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.Equal(t, "auth-code-123", r.Form.Get("code"))
		assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", r.Form.Get("client_assertion_type"))
		assert.NotEmpty(t, r.Form.Get("client_assertion"))

		resp := TokenResponse{
			AccessToken:  "mock-access-token-001",
			RefreshToken: "mock-refresh-token-001",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			Scope:        "accounts openid",
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	signer := setupTestSigner(t)
	client := NewConsentClient(mockServer.Client(), signer, mockServer.URL, mockServer.URL+"/token", "test-client")

	tokens, err := client.ExchangeCodeForToken(context.Background(), "auth-code-123", "https://app.midaz.local/callback")

	assert.NoError(t, err)
	assert.NotNil(t, tokens)
	assert.Equal(t, "mock-access-token-001", tokens.AccessToken)
	assert.Equal(t, 3600, tokens.ExpiresIn)
}
