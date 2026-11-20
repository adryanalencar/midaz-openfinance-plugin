# Midaz Plugin: Open Finance

An enterprise-grade plugin for the **[Midaz Ledger](https://github.com/LerianStudio/midaz)** that synchronizes real-world banking statements into immutable, balanced double-entry ledger transactions.

The plugin supports dual operational modes:
1. **Pluggy Aggregator (Recommended)**: Immediate bank connectivity across Brazilian institutions without requiring regulatory accreditation or ICP-Brasil certificates. Features frontend Connect Token generation and real-time event-driven webhook ingestion.
2. **Open Finance Brasil Direct (BACEN FAPI 1.0 Advanced)**: Direct regulated connectivity with banks via mutual TLS (mTLS), JWS payload signing (`PS256`), Open Finance Directory discovery, and BACEN-approved OAuth 2.0 flows.

---

## Table of Contents

- [Core Capabilities](#core-capabilities)
- [Double-Entry Accounting Architecture](#double-entry-accounting-architecture)
- [Operational Architecture](#operational-architecture)
- [Database Governance & Migrations](#database-governance--migrations)
- [Setup & Configuration](#setup--configuration)
  - [Prerequisites](#prerequisites)
  - [Option A: Setup with Pluggy Aggregator](#option-a-setup-with-pluggy-aggregator-recommended)
  - [Option B: Setup with Direct Open Finance Brasil](#option-b-setup-with-direct-open-finance-brasil)
  - [Database Configuration: Discrete vs DATABASE\_URL](#database-configuration-discrete-vs-database_url)
- [Running the Plugin](#running-the-plugin)
  - [Using Docker Compose (Recommended)](#using-docker-compose-recommended)
  - [Running Natively (Go)](#running-natively-go)
- [API Reference](#api-reference)
  - [Health & Manifest](#health--manifest)
  - [Pluggy Onboarding & Webhooks](#pluggy-onboarding--webhooks)
  - [Manual Sync Operations](#manual-sync-operations)
- [Testing & Quality Assurance](#testing--quality-assurance)
- [Project Structure](#project-structure)
- [Documentation Index](#documentation-index)

---

## Core Capabilities

- **Strict V1 Scope (Read-Only Mirroring):** Focuses entirely on bank statement line mirroring and continuous ledger synchronization. No payment initiation risk in V1.
- **Double-Entry Balance Guarantees:** Every bank deposit or withdrawal is balanced against a dedicated external clearing account (`@external/BRL:openfinance`).
- **Card Receivables Cross-Reconciliation:** Native inter-plugin protocol with `midaz-plugin-receivables` and **Matcher 6**. Accounts flagged with `reconciliation_enabled: true` coordinate card acquirer deposits to be settled from accrued receivables wallets rather than external clearing, eliminating the double-credit hazard.
- **Multi-Tenant Mother Account Hierarchy:** Supports structuring mirrored bank accounts as direct subaccounts under a merchant's Master CNPJ profile (`parentAccountId = mother_account_id`), mapped 1:1 to dedicated Matcher 6 tenants.
- **Mathematical Decimal Precision:** Zero floating-point rounding errors. Uses fixed-point scaling (`shopspring/decimal`) converted to integer amounts with `scale: 2`.
- **Real-Time Webhook Ingestion:** Ingests statements instantaneously upon Pluggy webhook notifications (`transactions/created`) using modern v2 cursor-based pagination (`/v2/transactions`).
- **Autonomous Background Worker:** Periodic fallback worker (`SyncWorker`) ensuring statement continuity for all active connections even if webhooks fail.
- **Idempotency & Deduplication:** Prevents duplicate postings in Midaz using deterministic transaction codes (`OFP-<txId>` / `OFP-PLG-<txId>`) and persistent hash indexing in PostgreSQL (`processed_transactions`).
- **Automated Database Governance:** Embedded `golang-migrate` engine runs on application boot with PostgreSQL advisory locks, guaranteeing schema synchronization across single or multi-replica deployments.
- **Midaz Console Dynamic Discovery:** Exposes `GET /api/manifest` allowing the plugin to mount dynamically inside the Midaz Console frontend.

---

## Double-Entry Accounting Architecture

In financial accounting with Midaz, all transactions must satisfy the fundamental equation:

$$\sum \text{Debits} = \sum \text{Credits}$$

To reflect real-world bank statements where money enters or leaves the banking institution, the plugin provisions a dedicated clearing account:
* **Account Alias:** `@external/BRL:openfinance`
* **Account Type:** `external` *(allows negative liability balances representing external liquidity)*

```mermaid
flowchart LR
    subgraph Inflow["Bank Inflow (CREDITO - Deposit / Pix In)"]
        direction LR
        ExtIn["@external/BRL:openfinance\n(Source - Debit)"] -->|Amount: R$ 1.500,00| MidazIn["@user/checking_account\n(Destination - Credit)"]
    end

    subgraph Outflow["Bank Outflow (DEBITO - Payment / Pix Out)"]
        direction LR
        MidazOut["@user/checking_account\n(Source - Debit)"] -->|Amount: R$ 250,00| ExtOut["@external/BRL:openfinance\n(Destination - Credit)"]
    end
```

### Amount Scaling & Serialization
Banking APIs provide string representations like `"1250.50"`. The plugin normalizes values to Midaz integer scale:
$$\text{Scaled Amount} = \text{Round}(\text{Decimal} \times 10^{\text{scale}}) \implies 1250.50 \times 10^2 = 125050$$

---

## Operational Architecture

Built with **Hexagonal Architecture (Ports and Adapters)** and **CQRS**:

```mermaid
flowchart TD
    subgraph External["External World"]
        Widget["Frontend / Pluggy Widget"]
        Webhook["Pluggy Webhook Dispatcher"]
        Banks["Open Finance Brasil Banks"]
        MidazCore["Midaz Ledger Core"]
    end

    subgraph HTTPIn["Inbound Adapters (Fiber HTTP)"]
        Router["Fiber Router (:8085)"]
        H_Health["/health"]
        H_Manifest["/api/manifest"]
        H_Pluggy["/v1/connect-token/pluggy\n/v1/connections/pluggy"]
        H_Webhook["/webhooks/pluggy"]
        H_Sync["/v1/sync\n/v1/sync/status"]
    end

    subgraph DomainCore["Domain & Application Services (CQRS)"]
        UC_PluggyConn["CreatePluggyConnectionUseCase"]
        UC_Webhook["HandlePluggyWebhookUseCase"]
        UC_Sync["SyncTransactionsUseCase"]
        UC_Status["GetSyncStatusUseCase"]
        Worker["Autonomous SyncWorker (Cron)"]
    end

    subgraph AdaptersOut["Outbound Adapters"]
        Storage["PostgreSQL Storage\n(bank_connections, processed_transactions)"]
        P_Pluggy["Pluggy Client v2\n(Cursor paging, Connect Token)"]
        P_OFDirect["Open Finance Direct\n(mTLS, JWS PS256, FAPI)"]
        MidazClient["Midaz Client\n(Double-Entry Posting)"]
    end

    Widget -->|Get Connect Token| H_Pluggy
    Widget -->|Register Item| H_Pluggy
    Webhook -->|transactions/created| H_Webhook

    H_Pluggy --> UC_PluggyConn
    H_Webhook --> UC_Webhook
    H_Sync --> UC_Sync
    Worker --> UC_Sync

    UC_PluggyConn --> Storage
    UC_Webhook --> Storage
    UC_Webhook --> P_Pluggy
    UC_Webhook --> MidazClient
    UC_Sync --> Storage
    UC_Sync --> P_Pluggy
    UC_Sync --> P_OFDirect
    UC_Sync --> MidazClient

    MidazClient -->|POST /transactions/json| MidazCore
```

---

## Database Governance & Migrations

The plugin manages its own PostgreSQL schema independently using `golang-migrate` embedded directly into the Go binary:

1. **Advisory Locks:** On application boot, the plugin acquires a PostgreSQL advisory lock (`migrate_lock`) to avoid race conditions when multiple container replicas start concurrently.
2. **Automatic Execution:** Controlled by `DB_AUTO_MIGRATE=true`.
3. **Tracking Table:** Uses `schema_migrations` matching the standard Midaz migration pattern.
4. **Database Migrations:**
   - `000001_init_schema`: Base tables for `bank_connections`, `processed_transactions`, and `sync_checkpoints`.
   - `000002_add_reconciliation_and_subaccounts`: Adds Mother Account hierarchy (`mother_account_id`, `mother_account_alias`), reconciliation controls (`reconciliation_enabled`, `reconciliation_receivables_account`, `matcher_context_id`), and transaction reconciliation status tracking (`status` with `POSTED_DIRECT` vs `SETTLED_VIA_RECEIVABLES`).

---

## Setup & Configuration

### Prerequisites
- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.24+](https://golang.org/) (if running natively)
- A running [Midaz](https://github.com/LerianStudio/midaz) instance

---

### Option A: Setup with Pluggy Aggregator (Recommended)

1. Copy the pre-configured Pluggy environment template:
   ```bash
   cp examples/.env.pluggy.example .env
   ```

2. Open `.env` and fill in your credentials from the [Pluggy Dashboard](https://dashboard.pluggy.ai):
   ```env
   PROVIDER=pluggy
   PLUGGY_BASE_URL=https://api.pluggy.ai
   PLUGGY_CLIENT_ID=your_pluggy_client_id
   PLUGGY_CLIENT_SECRET=your_pluggy_client_secret

   # Midaz Core Target
   MIDAZ_ONBOARDING_URL=http://localhost:3000
   MIDAZ_TRANSACTION_URL=http://localhost:3001
   MIDAZ_ORGANIZATION_ID=your-organization-uuid
   MIDAZ_LEDGER_ID=your-ledger-uuid
   ```

---

### Option B: Setup with Direct Open Finance Brasil

1. Copy the Direct Open Finance environment template:
   ```bash
   cp examples/.env.openfinance.example .env
   ```

2. Place your ICP-Brasil certificates in `./config/certs/` and update `.env`:
   ```env
   PROVIDER=openfinance
   OPENFINANCE_ENVIRONMENT=sandbox
   OPENFINANCE_DIRECTORY_URL=https://matls-auth.sandbox.directory.openbankingbrasil.org.br
   OPENFINANCE_CLIENT_ID=your_participant_client_id

   OPENFINANCE_TRANSPORT_CERT_PATH=./config/certs/transport.pem
   OPENFINANCE_TRANSPORT_KEY_PATH=./config/certs/transport.key
   OPENFINANCE_SIGNING_CERT_PATH=./config/certs/signing.pem
   OPENFINANCE_SIGNING_KEY_PATH=./config/certs/signing.key
   ```

---

### Database Configuration: Discrete vs DATABASE_URL

The plugin supports two flexible database connection modes:

* **Mode 1: Cloud Connection URI (`DATABASE_URL` / `DB_URI`)**
  If connecting to an external database (AWS RDS, Supabase, Neon, Cloud SQL), simply provide:
  ```env
  DATABASE_URL=postgres://midaz:secret@my-rds-instance.aws.com:5432/plugin_openfinance?sslmode=require
  ```
* **Mode 2: Discrete Parameters (Local / Docker)**
  ```env
  DB_HOST=localhost
  DB_PORT=5432
  DB_USER=midaz
  DB_PASSWORD=midaz_secure_password
  DB_NAME=plugin_openfinance
  DB_SSLMODE=disable
  ```

---

## Running the Plugin

### Using Docker Compose (Recommended)

Start the plugin along with its dedicated PostgreSQL database:
```bash
docker compose up -d
```

View container logs:
```bash
docker compose logs -f plugin-openfinance
```

Check health:
```bash
curl http://localhost:8085/health
```

---

### Running Natively (Go)

1. Ensure PostgreSQL is running on port 5432 (or configure `DATABASE_URL` in `.env`).
2. Run the application:
   ```bash
   go run cmd/app/main.go
   ```

---

## API Reference

### Health & Manifest

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Verifies service, Midaz ledger, and banking provider connectivity. |
| `GET` | `/api/manifest` | Exposes metadata for dynamic mounting in the Midaz Console. |

#### Example: `GET /health`
```json
{
  "status": "UP",
  "provider": "pluggy",
  "midaz_connected": true,
  "provider_connected": true,
  "timestamp": "2026-11-19T13:00:00Z"
}
```

---

### Pluggy Onboarding & Webhooks

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/connect-token/pluggy` | Generates a temporary Connect Token to open the Pluggy Connect Widget. |
| `POST` | `/v1/connections/pluggy` | Registers a newly authorized Item and maps account IDs to Midaz aliases. |
| `POST` | `/webhooks/pluggy` | Ingestion endpoint receiving Pluggy event webhooks. |

#### 1. Generate Connect Token
```bash
curl -X POST http://localhost:8085/v1/connect-token/pluggy \
  -H "Content-Type: application/json" \
  -d '{"clientUserId": "user-123"}'
```
Response:
```json
{
  "connectToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### 2. Register Connection & Account Mapping
After the user completes the Pluggy Widget on the frontend, register the connection:
```bash
curl -X POST http://localhost:8085/v1/connections/pluggy \
  -H "Content-Type: application/json" \
  -d '{
    "itemId": "e6a0d4b8-2bc1-4e89-8d77-d3527b13e9a1",
    "accountId": "a3b9f4e2-9d8a-4c7b-8e1f-2a3b4c5d6e7f",
    "midazAccountAlias": "@company/itau_checking"
  }'
```
Response:
```json
{
  "status": "success",
  "connectionId": "550e8400-e29b-41d4-a716-446655440000",
  "itemId": "e6a0d4b8-2bc1-4e89-8d77-d3527b13e9a1",
  "accountId": "a3b9f4e2-9d8a-4c7b-8e1f-2a3b4c5d6e7f",
  "midazAccountAlias": "@company/itau_checking",
  "bankCompe": "341"
}
```

#### 3. Real-Time Webhook (`POST /webhooks/pluggy`)
Pluggy dispatches events such as:
- `transactions/created`: Pulls `createdTransactionsLink` cursor, converts amounts to scale 2, and creates double-entry transactions in Midaz.
- `item/updated`, `item/error`, `item/waiting_user_input`, `item/deleted`: Updates connection lifecycle status in PostgreSQL.

---

### Manual Sync Operations

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/sync` | Triggers immediate statement reconciliation for a bank account. |
| `GET` | `/v1/sync/status` | Returns sync execution statistics, metrics, and provider status. |

#### Trigger Manual Sync:
```bash
curl -X POST http://localhost:8085/v1/sync \
  -H "Content-Type: application/json" \
  -d '{
    "bank_account_id": "a3b9f4e2-9d8a-4c7b-8e1f-2a3b4c5d6e7f",
    "from_date": "2026-11-01",
    "to_date": "2026-11-19"
  }'
```

---

### Inter-Plugin Coordination (Receivables Handshake)

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/v1/internal/transactions/mark-processed` | Internal endpoint invoked by `midaz-plugin-receivables` when a bank transaction is settled via Matcher 6 reconciliation. |

#### Handshake Payload:
```json
{
  "bankTransactionId": "OFP-SICREDI-998877",
  "midazTransactionCode": "MIDAZ-MULTI-777",
  "status": "SETTLED_VIA_RECEIVABLES"
}
```
* **Behavior:** Idempotently records the transaction ID in `processed_transactions` with status `SETTLED_VIA_RECEIVABLES`, ensuring Open Finance bypasses direct double-entry creation and prevents double-crediting the merchant's account.

---

## Testing & Quality Assurance

The plugin enforces unit and integration testing across all adapters and business use cases using `testify` and `go-sqlmock`:

Run tests locally:
```bash
go test -v ./...
```

Run tests inside an isolated Docker container:
```bash
docker run --rm -v "$(pwd):/app" -w /app golang:1.24-alpine go test -v ./...
```

Build the production multi-stage Docker container:
```bash
docker build -t midaz-plugin-openfinance:latest .
```

---

## Project Structure

```
plugin-openfinance/
├── cmd/
│   └── app/
│       └── main.go                 # Application bootstrap & dependency injection
├── internal/
│   ├── adapters/
│   │   ├── http/
│   │   │   ├── in/                 # Fiber HTTP handlers, router & manifest
│   │   │   └── out/
│   │   │       └── midaz/          # Midaz double-entry posting & clearing account client
│   │   ├── postgres/               # PostgreSQL repository & embedded golang-migrate runner
│   │   └── providers/
│   │       ├── openfinance/        # Direct BACEN FAPI 1.0 mTLS & JWS engine
│   │       └── pluggy/             # Pluggy v2 client, cursor pagination & webhooks
│   ├── bootstrap/                  # Config loader, PostgreSQL DSN builder & server lifecycle
│   ├── domain/                     # Core domain entities (BankTransaction, BankConnection)
│   ├── services/
│   │   ├── command/                # CQRS Commands (Sync, Webhooks, Onboarding)
│   │   └── query/                  # CQRS Queries (Sync status, metrics)
│   └── workers/                    # Autonomous background sync worker
├── migrations/                     # Embedded SQL schema migrations (000001_init_schema)
├── examples/                       # Pre-configured environment profiles
│   ├── .env.pluggy.example         # Pluggy Aggregator profile
│   └── .env.openfinance.example    # Direct Open Finance Brasil profile
├── docs/                           # Deep-dive architecture and design specifications
├── Dockerfile                      # Multi-stage production container
└── docker-compose.yml              # Local orchestration with companion PostgreSQL
```

---

## Documentation Index

For deep-dive architectural specifications, refer to the documents in [`docs/`](docs/):

* [01. Product Requirements Document](docs/01-PRODUCT-REQUIREMENTS.md)
* [02. Technical Architecture Specification](docs/02-TECHNICAL-ARCHITECTURE.md)
* [03. Double-Entry & Midaz Accounting Model](docs/03-ACCOUNTING-DOUBLE-ENTRY.md)
* [04. Open Finance Brasil Direct (BACEN FAPI) Specification](docs/04-OPEN-FINANCE-BRASIL-SPEC.md)
* [05. Multi-Provider Strategy & Aggregator Fallback](docs/05-PROVIDER-STRATEGY.md)
* [06. API & Midaz Console Manifest Specification](docs/06-API-AND-MANIFEST-SPEC.md)
* [07. Implementation Plan & Traceability Matrix](docs/07-IMPLEMENTATION-PLAN.md)
* [08. Git Rules & Conventional Commits](docs/08-GIT-RULES-AND-CONVENTIONS.md)
* [09. Testing Strategy & Test Suites](docs/09-TESTING-STRATEGY.md)
* [10. Pluggy Provider Architecture Specification](docs/10-PLUGGY-PROVIDER-SPEC.md)
