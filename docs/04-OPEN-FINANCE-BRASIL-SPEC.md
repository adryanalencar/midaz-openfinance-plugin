# Open Finance Brasil Direct Integration Specification

## 1. Regulatory Context & References

Direct integration with Brazilian financial institutions adheres to the normative framework established by the Central Bank of Brazil (BACEN) and the Open Finance Brasil Governance Structure:
- **OpenAPI Repositories:** [OpenBanking-Brasil/openapi](https://github.com/OpenBanking-Brasil/openapi)
- **Security Profile:** Financial-grade API (FAPI) 1.0 Advanced / FAPI 2.0 Security Profile.
- **Participant Directory:** Open Finance Brasil Directory of Participants (Sandbox & Production).

---

## 2. Cryptographic & Security Architecture

Direct communication with bank APIs requires dual ICP-Brasil Digital Certificates:

```
                  ┌──────────────────────────────────────────────┐
                  │          ICP-Brasil Certificates             │
                  ├──────────────────────┬───────────────────────┤
                  │  Transport (mTLS)    │    Signature (JWS)    │
                  ├──────────────────────┼───────────────────────┤
                  │ Used for TLS channel │ Used for signing HTTP │
                  │ authentication with  │ request bodies & JWS  │
                  │ banking AS and RS    │ consent payloads      │
                  └──────────────────────┴───────────────────────┘
```

### 2.1. Mutual TLS (mTLS)
- All connections to Authorization Servers (AS) and Resource Servers (RS) must enforce TLS 1.2 or 1.3 using cipher suites mandated by BACEN.
- The client presents the ICP-Brasil Transport Certificate (`OPENFINANCE_CERT_TRANSPORT_PEM` and `OPENFINANCE_KEY_TRANSPORT_PEM`).

### 2.2. Client Authentication: `private_key_jwt`
- Client authentication at the OAuth 2.0 `/token` endpoint uses a signed JWT (`urn:ietf:params:oauth:client-assertion-type:jwt-bearer`).
- The JWT is signed using the ICP-Brasil Signature private key via `PS256` or `ES256`.

---

## 3. Core API Workflows for Transaction Mirroring

### 3.1. Phase 1: Consent Creation (`/consents/v3/consents`)
Before reading transactions, the customer must authorize access to their accounts.

* **Endpoint:** `POST /open-banking/consents/v3/consents`
* **Headers:**
  - `Authorization: Bearer <client_credentials_token>`
  - `x-fapi-interaction-id: <uuid>`
  - `x-fapi-auth-date: <rfc7231-date>`
  - `x-idempotency-key: <uuid>`
* **Payload:**
```json
{
  "data": {
    "permissions": [
      "ACCOUNTS_READ",
      "ACCOUNTS_BALANCES_READ",
      "ACCOUNTS_TRANSACTIONS_READ"
    ],
    "expirationDateTime": "2027-11-19T00:00:00Z",
    "transactionFromDateTime": "2025-11-19T00:00:00Z",
    "transactionToDateTime": "2026-11-19T23:59:59Z"
  }
}
```
* **Response Status:** `201 Created` with `status: "AWAITING_AUTHORISATION"`.

---

### 3.2. Phase 2: User Authorization & Token Exchange
1. The user is redirected to the bank's authorization URL with:
   - `response_type=code id_token`
   - `scope=openid consent:<consentId>`
   - `client_id=<registered_client_id>`
   - `request=<signed_request_object_jwt>`
2. The user authenticates and approves on their bank app.
3. The bank redirects back with an authorization `code`.
4. The plugin exchanges the code for an `access_token` and `refresh_token` at `/token`.

---

### 3.3. Phase 3: Accounts Listing (`/accounts/v2/accounts`)
* **Endpoint:** `GET /open-banking/accounts/v2/accounts`
* **Response:**
```json
{
  "data": [
    {
      "brandName": "Banco Itaú",
      "companyCnpj": "60701190000104",
      "type": "CONTA_DEPOSITO_A_VISTA",
      "compeCode": "341",
      "branchCode": "0001",
      "number": "12345678",
      "checkDigit": "9",
      "accountId": "018f9211-5b72-74ba-9132-e018d9f1092a"
    }
  ]
}
```

---

### 3.4. Phase 4: Transaction Ingestion (`/accounts/v2/accounts/{accountId}/transactions`)
* **Endpoint:** `GET /open-banking/accounts/v2/accounts/{accountId}/transactions`
* **Query Parameters:**
  - `fromBookingDate`: Filter starting date (`YYYY-MM-DD`).
  - `toBookingDate`: Filter end date (`YYYY-MM-DD`).
  - `page`: Page number (default: 1).
  - `page-size`: Records per page (maximum: 1000).
* **Response Payload Example:**
```json
{
  "data": [
    {
      "transactionId": "TX-ITAU-20261119-001239",
      "completedAuthorisedPaymentType": "PIX",
      "creditDebitType": "CREDITO",
      "transactionName": "PIX RECEBIDO - SERVICOS TECH",
      "type": "PIX",
      "amount": "2500.00",
      "transactionCurrency": "BRL",
      "transactionDate": "2026-11-19",
      "bookingDateTime": "2026-11-19T10:15:30Z"
    },
    {
      "transactionId": "TX-ITAU-20261119-001240",
      "completedAuthorisedPaymentType": "BOLETO",
      "creditDebitType": "DEBITO",
      "transactionName": "PAGTO ELETRON COBRANCA",
      "type": "BOLETO",
      "amount": "142.30",
      "transactionCurrency": "BRL",
      "transactionDate": "2026-11-19",
      "bookingDateTime": "2026-11-19T11:42:00Z"
    }
  ],
  "meta": {
    "totalRecords": 2,
    "totalPages": 1
  }
}
```

---

## 4. Rate Limits & Exponential Backoff

1. **Throttling Policies:** Banks enforce per-second and per-minute rate limits (typically 10-30 req/sec per consent).
2. **HTTP 429 Handling:** When a `429 Too Many Requests` response is received, the client inspects the `Retry-After` header and applies exponential backoff with full jitter:
   $$T_{\text{sleep}} = \min(T_{\text{max}}, T_{\text{base}} \times 2^{\text{attempt}}) + \text{random\_jitter}$$
3. **Daily Quotas:** Routine sync intervals default to once every 1 or 2 hours, preventing unnecessary quota exhaustion.
