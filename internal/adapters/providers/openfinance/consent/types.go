package consent

import "time"

// Mandatory permissions for account statement and balance synchronization according to Open Finance Brasil
var StandardSyncPermissions = []string{
	"ACCOUNTS_READ",
	"ACCOUNTS_BALANCES_READ",
	"ACCOUNTS_TRANSACTIONS_READ",
	"RESOURCES_READ",
}

// CreateConsentRequest matches the schema defined in consents-3.3.1.yml.
type CreateConsentRequest struct {
	Data ConsentRequestData `json:"data"`
}

type ConsentRequestData struct {
	Permissions             []string `json:"permissions"`
	ExpirationDateTime      string   `json:"expirationDateTime"`
	TransactionFromDateTime string   `json:"transactionFromDateTime"`
	TransactionToDateTime   string   `json:"transactionToDateTime"`
}

// ConsentResponse matches the return schema for created consents.
type ConsentResponse struct {
	Data  ConsentResponseData `json:"data"`
	Links map[string]string   `json:"links,omitempty"`
	Meta  map[string]any      `json:"meta,omitempty"`
}

type ConsentResponseData struct {
	ConsentID          string    `json:"consentId"`
	CreationDateTime   time.Time `json:"creationDateTime"`
	Status             string    `json:"status"` // AWAITING_AUTHORISATION, AUTHORISED, REJECTED, REVOKED
	StatusUpdateDate   time.Time `json:"statusUpdateDateTime"`
	Permissions        []string  `json:"permissions"`
	ExpirationDateTime string    `json:"expirationDateTime"`
}

// TokenResponse matches standard OAuth2 tokens returned by bank Authorization Servers.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}
