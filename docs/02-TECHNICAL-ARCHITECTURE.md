# Technical Architecture & Design: Midaz Open Finance Plugin

## 1. System Overview & Boundaries

The **Midaz Open Finance Plugin** is built as an independent microservice that strictly adheres to the architectural, coding, and operational standards established in the Midaz Core repository (`github.com/LerianStudio/midaz/v3`).

```mermaid
flowchart TB
    subgraph ExternalEcosystem["External Financial Ecosystem"]
        direction TB
        OFP["Open Finance Brasil APIs<br/>(FAPI 1.0 Advanced + mTLS)"]
        PLG["Aggregator API<br/>(e.g., Pluggy)"]
    end

    subgraph PluginService["Midaz Plugin: Open Finance (Go + Fiber)"]
        direction TB
        Main["cmd/app/main.go"] --> Bootstrap["internal/bootstrap"]
        
        subgraph PortsAdapters["Hexagonal Architecture"]
            HTTPIn["Adapters: HTTP In<br/>• /api/manifest<br/>• /health<br/>• /v1/sync"]
            
            subgraph AppServices["Application Services (CQRS)"]
                CmdSync["Command: SyncTransactions"]
                CmdBootstrap["Command: EnsureExternalAccount"]
                QryStatus["Query: GetSyncStatus"]
                QryManifest["Query: GetManifest"]
            end
            
            subgraph Domain["Domain Layer"]
                IProvider["Port: OpenFinanceProvider"]
                IMidaz["Port: MidazClient"]
                ISyncState["Port: SyncStateRepository"]
            end
            
            subgraph ExtAdapters["Adapters: Out"]
                AdapterOFP["Provider: Direct Open Finance"]
                AdapterPLG["Provider: Pluggy Aggregator"]
                AdapterMidaz["HTTP Out: Midaz Client"]
                AdapterDB["Storage: PostgreSQL / SQLite"]
            end
        end

        HTTPIn --> AppServices
        AppServices --> Domain
        IProvider -.-> AdapterOFP
        IProvider -.-> AdapterPLG
        IMidaz -.-> AdapterMidaz
        ISyncState -.-> AdapterDB
    end

    subgraph MidazCore["Midaz Core Engine"]
        Onboarding["Midaz Onboarding API<br/>Port 3000"]
        Transaction["Midaz Transaction API<br/>Port 3001"]
        Console["Midaz Console<br/>(Next.js Microfrontend)"]
    end

    AdapterOFP -->|mTLS + FAPI| OFP
    AdapterPLG -->|HTTPS REST| PLG
    AdapterMidaz -->|REST /v1/...| Onboarding
    AdapterMidaz -->|REST /v1/...| Transaction
    Console -->|Fetch Manifest & Remote UI| HTTPIn
```

---

## 2. Clean Architecture & CQRS Implementation

To ensure seamless maintainability and code parity with Midaz core components (`components/onboarding` and `components/transaction`), the plugin adopts the Hexagonal and CQRS patterns:

### 2.1. Directory Structure

```
plugins/plugin-openfinance/
├── cmd/
│   └── app/
│       └── main.go                  # Service entrypoint, env initialization, lifecycle
├── internal/
│   ├── bootstrap/
│   │   ├── config.go                # Top-level configuration struct with env tags
│   │   └── service.go               # Dependency injection container & server runner
│   ├── domain/
│   │   ├── provider.go              # OpenFinanceProvider interface & transaction contracts
│   │   ├── account_mapping.go       # Bank account to Midaz account association
│   │   └── sync_state.go            # Checkpoint metadata & idempotency records
│   ├── services/
│   │   ├── command/
│   │   │   ├── sync_transactions.go # Orchestrates batch pulling and ledger posting
│   │   │   └── ensure_account.go    # Validates and creates @external/BRL account
│   │   └── query/
│   │       ├── get_manifest.go      # Provides Midaz Console manifest
│   │       └── get_sync_status.go   # Queries sync health and execution history
│   └── adapters/
│       ├── http/
│       │   └── in/
│       │       ├── router.go        # Fiber web router setup & middleware attachment
│       │       ├── manifest.go      # /api/manifest handler
│       │       ├── sync_handler.go  # Manual sync trigger handler
│       │       └── health.go        # Health check handler
│       ├── http/
│       │   └── out/
│       │       └── midaz/
│       │           └── client.go    # Midaz HTTP client (account lookup, transaction posting)
│       ├── providers/
│       │   ├── factory.go           # Instantiates provider based on cfg.Provider
│       │   ├── openfinance/
│       │   │   └── client.go        # Direct Open Finance Brasil (mTLS, JWS, FAPI)
│       │   └── pluggy/
│       │       └── client.go        # Pluggy aggregator client
│       └── storage/
│           └── postgres/            # Checkpoint and idempotency persistence
```

---

## 3. End-to-End Sequence Diagrams

### 3.1. Startup & Clearing Account Bootstrapping

Before processing any transaction, the plugin ensures that the ledger has the appropriate `@external` clearing account registered.

```mermaid
sequenceDiagram
    autonumber
    participant App as Plugin Bootstrap
    participant MidazOnboarding as Midaz Onboarding API
    participant MidazDB as Midaz Ledger DB

    App->>App: Load Env Config (PROVIDER, MIDAZ_API_URL, LEDGER_ID)
    App->>MidazOnboarding: GET /v1/organizations/{org_id}/ledgers/{ledger_id}/accounts?alias=@external/BRL:openfinance
    alt Account exists
        MidazOnboarding-->>App: 200 OK (Account Details)
    else Account not found (404)
        App->>MidazOnboarding: POST /v1/organizations/{org_id}/ledgers/{ledger_id}/accounts
        Note over App,MidazOnboarding: Body: { name: "Open Finance Clearing", assetCode: "BRL", type: "external", alias: "@external/BRL:openfinance" }
        MidazOnboarding->>MidazDB: Insert Account with type="external"
        MidazOnboarding-->>App: 201 Created
    end
    App->>App: Start Background Sync Worker & HTTP Server
```

---

### 3.2. Transaction Sync and Idempotent Double-Entry Ingestion

```mermaid
sequenceDiagram
    autonumber
    participant Worker as Sync Worker (Plugin)
    participant Provider as OpenFinanceProvider (Direct/Pluggy)
    participant BankAPI as Bank API
    participant MidazTx as Midaz Transaction API
    participant Storage as Local Checkpoint Storage

    Worker->>Storage: GetLastSyncCheckpoint(bankAccountId)
    Storage-->>Worker: lastBookingDate = "2026-11-18"

    Worker->>Provider: FetchTransactions(bankAccountId, fromDate, toDate)
    Provider->>BankAPI: GET /accounts/{id}/transactions?fromBookingDate=2026-11-18
    BankAPI-->>Provider: [ { id: "TX100", type: "CREDITO", amount: "50.00", name: "Pix Recebido" } ]
    Provider-->>Worker: Normalized Domain Transactions

    loop For each transaction
        Worker->>Storage: IsTransactionProcessed("TX100")
        alt Not Processed
            Worker->>Worker: Build Double-Entry Payload:<br/>Source: @external/BRL:openfinance<br/>Dest: Account_Target<br/>Code: OFP-TX100
            Worker->>MidazTx: POST /v1/organizations/{org_id}/ledgers/{ledger_id}/transactions/json
            alt 201 Created or 200 OK
                MidazTx-->>Worker: Transaction Posted (Balance Updated)
                Worker->>Storage: MarkTransactionProcessed("TX100", timestamp)
            else 409 Conflict (Duplicate code in Midaz)
                Worker->>Storage: MarkTransactionProcessed("TX100", timestamp)
            end
        else Already Processed
            Note over Worker: Skip transaction (Idempotent)
        end
    end

    Worker->>Storage: UpdateLastSyncCheckpoint(bankAccountId, newCheckpointDate)
```

---

## 4. Concurrency & Idempotency Guarantees

1. **Deterministic Unique Codes:**
   Every transaction pushed to Midaz has its `code` formatted as:
   `code = "OFP-" + providerTransactionId`
   Midaz enforces unique constraint indexing on `(ledger_id, code)`. If a concurrent worker or network retry triggers an identical post, Midaz rejects the second insert with a duplicate code error, preserving ledger integrity.

2. **Distributed Locking:**
   When running multiple replicas of the plugin, the `SyncWorker` acquires a Redis/PostgreSQL advisory lock per `bank_account_id` before starting an ingestion cycle to eliminate race conditions.

3. **Safe Lookback Window:**
   Bank statement APIs occasionally back-date or settle transactions after the fact. The sync worker always requests transactions starting from `lastBookingDate - 1 day` (a 24-hour lookback window), relying on the idempotency engine to discard previously ingested entries.
