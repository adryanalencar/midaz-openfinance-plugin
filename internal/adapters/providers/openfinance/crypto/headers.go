package crypto

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// FAPIHeaderConstants defines the standard HTTP header names mandated by Open Finance Brasil.
const (
	HeaderXFapiInteractionID     = "x-fapi-interaction-id"
	HeaderXFapiAuthDate          = "x-fapi-auth-date"
	HeaderXFapiCustomerIPAddress = "x-fapi-customer-ip-address"
	HeaderXCustomerUserAgent     = "x-customer-user-agent"
	HeaderIdempotencyKey         = "x-idempotency-key"
)

// InjectFAPIHeaders populates the required FAPI 1.0 Advanced headers on outbound banking requests.
func InjectFAPIHeaders(req *http.Request, customerIP, userAgent string) string {
	interactionID := uuid.New().String()

	req.Header.Set(HeaderXFapiInteractionID, interactionID)
	req.Header.Set(HeaderXFapiAuthDate, time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("Accept", "application/json")

	if customerIP != "" {
		req.Header.Set(HeaderXFapiCustomerIPAddress, customerIP)
	}

	if userAgent != "" {
		req.Header.Set(HeaderXCustomerUserAgent, userAgent)
	} else {
		req.Header.Set(HeaderXCustomerUserAgent, "Midaz-OpenFinance-Plugin/v1.0")
	}

	return interactionID
}

// ExtractInteractionID retrieves the correlation interaction ID from response headers.
func ExtractInteractionID(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	return resp.Header.Get(HeaderXFapiInteractionID)
}
