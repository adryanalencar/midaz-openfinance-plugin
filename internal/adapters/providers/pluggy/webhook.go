package pluggy

// WebhookEvent represents an incoming event notification from Pluggy.
type WebhookEvent struct {
	Event                     string `json:"event"`
	EventID                   string `json:"eventId"`
	ID                        string `json:"id"`
	ItemID                    string `json:"itemId"`
	AccountID                 string `json:"accountId"`
	ClientID                  string `json:"clientId"`
	CreatedTransactionsLink   string `json:"createdTransactionsLink"`
	CreatedTransactionsLinkV2 string `json:"createdTransactionsLinkV2"`
	TransactionsCreatedAtFrom string `json:"transactionsCreatedAtFrom"`
	TriggeredBy               string `json:"triggeredBy"`
	ClientUserID              string `json:"clientUserId"`
}

// GetTransactionsLink returns the preferred v2 transactions URL from the webhook payload.
func (e *WebhookEvent) GetTransactionsLink() string {
	if e.CreatedTransactionsLinkV2 != "" {
		return e.CreatedTransactionsLinkV2
	}
	return e.CreatedTransactionsLink
}
