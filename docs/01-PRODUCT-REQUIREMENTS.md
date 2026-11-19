# Product Requirements Document (PRD): Midaz Open Finance Plugin

## 1. Problem Statement
Modern fintechs, corporate treasuries, and core banking platforms using Midaz as their source-of-truth ledger frequently hold operational bank accounts across multiple external financial institutions (e.g., Itaú, Bradesco, Banco do Brasil, Santander, Nubank). 

Today, reconciling transactions from these external accounts requires either manual CSV/OFX imports or fragmented proprietary integrations. There is no automated, real-time mechanism to mirror real-world banking statement lines into Midaz's immutable double-entry ledger while preserving mathematical invariants.

## 2. Product Objectives
The **Midaz Open Finance Plugin** provides automated, scheduled, and on-demand synchronization of transactional banking histories from external institutions directly into Midaz.

### Key Objectives:
- **Automate Bank Statement Mirroring:** Eliminate manual imports by continuously pulling bank transactions via Brazil's Open Finance standard or regulated aggregators.
- **Enforce Double-Entry Soundness:** Translate single-line bank statement entries (credits/debits) into balanced, auditable Midaz transactions using dedicated `@external` accounts.
- **Provider Agnostic Infrastructure:** Support both direct regulated Open Finance Brasil APIs (FAPI 1.0 Advanced + mTLS) and aggregators (Pluggy, Belvo) through simple configuration flags.
- **Seamless Console Integration:** Plug directly into the Midaz Console as a dynamic microfrontend plugin via `/api/manifest`.

---

## 3. Scope Definition

### 3.1. In-Scope for V1: Transactional Mirroring Only
- **External Account Discovery & Mapping:** Fetch accounts authorized by the user/company and link each bank account to a corresponding internal Midaz account.
- **Historical Transaction Ingestion:** Ingest historical transactions (`fromBookingDate` to `toBookingDate`) up to the maximum window supported by the institution (typically 12 months).
- **Continuous Incremental Sync:** Scheduled polling worker that queries delta transactions since the last successful sync checkpoint.
- **Idempotent Ledger Posting:** Ensure zero duplicates even when batches overlap, using the Open Finance `transactionId` as a unique ledger transaction code/metadata key.
- **Clearance Account Management:** Automatically verify and bind transactions to `@external/BRL:openfinance` (or default `@external/BRL`) to accommodate external cash inflows and outflows.
- **Provider Switchability:** Switch between direct Open Finance and aggregator mode using `PROVIDER=openfinance` or `PROVIDER=pluggy`.

### 3.2. Explicit Non-Goals (Out-of-Scope for V1)
- **Payment Initiation (ITP / Iniciação de Pagamentos):** V1 will **NOT** initiate payments, transfers, Pix, TEDs, or schedule payouts. The plugin is strictly **read-only** with respect to the banking institution.
- **Credit / Financing Products:** No ingestion of loans, credit cards, or investments for V1. Focus is solely on Current Accounts (*Conta Corrente*) and Savings Accounts (*Conta Poupança*).
- **Automated Settlement / Sweeping:** No automated balance rebalancing across bank accounts.

---

## 4. User Personas

| Persona | Role | Key Goal |
| :--- | :--- | :--- |
| **Treasury Officer** | Manages multi-bank liquidity and cash positions | Wants an aggregated, real-time view of all bank balances and transactions inside Midaz without logging into multiple bank portals. |
| **Financial Controller / Auditor** | Ensures balance sheet accuracy and regulatory compliance | Demands that every single cent deposited or withdrawn in external banks is balanced against an auditable clearing account in Midaz. |
| **Platform Engineer** | Operates and configures Midaz infrastructure | Needs a plug-and-play service deployable via Docker, configurable exclusively through environment variables, requiring zero hardcoded bank logic. |

---

## 5. Functional Requirements (FR)

### FR-01: Clearing Account Bootstrapping
- Upon startup, the plugin checks if the designated clearing account (e.g. `alias: @external/BRL:openfinance`) exists within the target Midaz Ledger.
- If absent, it calls Midaz Onboarding API (`POST /v1/organizations/{org_id}/ledgers/{ledger_id}/accounts`) with `type: "external"` to initialize it.

### FR-02: Multi-Provider Ingestion Engine
- The plugin must read `PROVIDER` from environment variables:
  - If `PROVIDER=openfinance`: Connects directly to bank endpoints using FAPI 1.0 mTLS and Open Finance Brasil OpenAPI specs.
  - If `PROVIDER=pluggy`: Connects via Pluggy API using `PLUGGY_CLIENT_ID` and `PLUGGY_CLIENT_SECRET`.
- Both implementations must produce identical internal domain transaction models.

### FR-03: Transaction Ingestion & Normalization
- Ingest transactions including:
  - `transactionId`: Unique ID from the banking institution.
  - `amount`: Monetary value in decimal string (e.g. `"150.25"`).
  - `creditDebitType`: `CREDITO` (inflow) or `DEBITO` (outflow).
  - `transactionName`: Human-readable statement description.
  - `bookingDateTime`: Exact accounting booking timestamp.
  - `paymentType`: Pix, TED, DOC, Boleto, Fee, etc.

### FR-04: Double-Entry Posting to Midaz
- For `CREDITO` (Bank Inflow / Deposit):
  - **Source:** `@external/BRL:openfinance`
  - **Destination:** Target Account in Midaz
- For `DEBITO` (Bank Outflow / Withdrawal):
  - **Source:** Target Account in Midaz
  - **Destination:** `@external/BRL:openfinance`
- Transaction code must be deterministic: `OFP-<transactionId>` to leverage Midaz's built-in unique constraint.

### FR-05: Sync Checkpoint & Resiliency
- Persist the `last_sync_timestamp` per connected account.
- When resuming, query transactions from `last_sync_timestamp - 1 day` (to safely capture delayed postings or settlement adjustments) and rely on idempotency deduplication.

---

## 6. Acceptance Criteria
1. Given an external bank account with a R$ 500 Pix deposit, the plugin posts a transaction to Midaz debiting `@external/BRL:openfinance` by 500 and crediting the Midaz account by 500.
2. Given a network failure during sync, subsequent sync runs do not create duplicate transactions in Midaz.
3. Switching `PROVIDER=pluggy` to `PROVIDER=openfinance` requires only changing environment variables and restarting the container.
4. The plugin serves `GET /api/manifest` returning valid Midaz Console Plugin metadata.
