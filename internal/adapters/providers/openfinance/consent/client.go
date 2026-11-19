package consent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/crypto"
)

// Client coordinates Open Finance consent operations and OAuth2 token management.
type Client struct {
	httpClient *http.Client
	signer     *crypto.JWSSigner
	baseURL    string
	tokenURL   string
	clientID   string
}

// NewConsentClient initializes a new ConsentClient.
func NewConsentClient(httpClient *http.Client, signer *crypto.JWSSigner, baseURL, tokenURL, clientID string) *Client {
	return &Client{
		httpClient: httpClient,
		signer:     signer,
		baseURL:    baseURL,
		tokenURL:   tokenURL,
		clientID:   clientID,
	}
}

// BuildStandardConsentPayload produces a compliant consent payload for up to 12 months historical window.
func BuildStandardConsentPayload(expiration time.Time) CreateConsentRequest {
	now := time.Now().UTC()
	oneYearAgo := now.AddDate(-1, 0, 0)

	return CreateConsentRequest{
		Data: ConsentRequestData{
			Permissions:             StandardSyncPermissions,
			ExpirationDateTime:      expiration.Format(time.RFC3339),
			TransactionFromDateTime: oneYearAgo.Format(time.RFC3339),
			TransactionToDateTime:   now.Format(time.RFC3339),
		},
	}
}

// CreateConsent submits POST /open-banking/consents/v3/consents to the bank API.
func (c *Client) CreateConsent(ctx context.Context, clientToken string, req CreateConsentRequest) (*ConsentResponse, error) {
	endpoint := fmt.Sprintf("%s/open-banking/consents/v3/consents", c.baseURL)

	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize consent request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	crypto.InjectFAPIHeaders(httpReq, "", "")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+clientToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("consent request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bank rejected consent creation: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var consentResp ConsentResponse
	if err := json.Unmarshal(respBody, &consentResp); err != nil {
		return nil, fmt.Errorf("failed to decode consent response: %w", err)
	}

	return &consentResp, nil
}

// GetConsent queries GET /open-banking/consents/v3/consents/{consentId}.
func (c *Client) GetConsent(ctx context.Context, clientToken, consentID string) (*ConsentResponse, error) {
	endpoint := fmt.Sprintf("%s/open-banking/consents/v3/consents/%s", c.baseURL, consentID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	crypto.InjectFAPIHeaders(httpReq, "", "")
	httpReq.Header.Set("Authorization", "Bearer "+clientToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve consent: status %d", resp.StatusCode)
	}

	var consentResp ConsentResponse
	if err := json.NewDecoder(resp.Body).Decode(&consentResp); err != nil {
		return nil, err
	}

	return &consentResp, nil
}

// ExchangeCodeForToken exchanges an authorization_code for OAuth2 tokens using private_key_jwt.
func (c *Client) ExchangeCodeForToken(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	assertion, err := c.signer.GenerateClientAssertion(c.tokenURL, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to sign client assertion: %w", err)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", c.clientID)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", assertion)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
