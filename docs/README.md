# Midaz Open Finance Plugin Documentation

Welcome to the official documentation for the **Midaz Open Finance Plugin** (`plugin-openfinance`).

This plugin connects external financial institutions to the **Midaz Ledger** through Brazil's Open Finance ecosystem and modern aggregation APIs (such as Pluggy), automatically syncing external bank transactions into Midaz with strict double-entry balance guarantees and `@external/` clearing accounts.

## Documentation Index

| Document | Description |
| :--- | :--- |
| [01. Product Requirements Document (PRD)](01-PRODUCT-REQUIREMENTS.md) | Business goals, V1 scope (transaction mirroring only), non-goals (no payment initiation), personas, and acceptance criteria. |
| [02. Technical Architecture & Design](02-TECHNICAL-ARCHITECTURE.md) | High-level system architecture, component diagrams, sequence diagrams, CQRS & Hexagonal structure following Midaz standards. |
| [03. Accounting & Double-Entry Specification](03-ACCOUNTING-DOUBLE-ENTRY.md) | Midaz ledger accounting model, `@external/` account semantics, inflow/outflow mapping, and mathematical balance invariants. |
| [04. Open Finance Brasil Direct Specification](04-OPEN-FINANCE-BRASIL-SPEC.md) | Direct integration details based on `OpenBanking-Brasil/openapi`: FAPI 1.0 Advanced, mTLS with ICP-Brasil certificates, DCR, Consents API v3, Accounts API v2, and Transactions API v2. |
| [05. Multi-Provider Strategy & Extensibility](05-PROVIDER-STRATEGY.md) | Pluggable provider design pattern configured via `PROVIDER` environment variable (`openfinance` direct vs. `pluggy` aggregator). |
| [06. API & Plugin Manifest Specification](06-API-AND-MANIFEST-SPEC.md) | HTTP endpoints, health checks, manual synchronization triggers, and the Midaz Console `/api/manifest` integration contract. |
| [07. Detailed Implementation Plan (Epics, Stories & Tasks)](07-IMPLEMENTATION-PLAN.md) | Complete backlog with 8 Epics, user stories, tasks, validation rules, and traceability matrix based directly on official Open Finance OpenAPI YAML specs. |
| [08. Git Rules, Branching Model & Commit Conventions](08-GIT-RULES-AND-CONVENTIONS.md) | Midaz GitFlow branching strategy, mandatory PR targeting rules (`develop`), branch naming conventions, Conventional Commits specification, and PR checklist. |
| [09. Testing Strategy & Quality Assurance Standards](09-TESTING-STRATEGY.md) | Test pyramid, table-driven testing pattern, Testify assertions, double-entry mathematical invariant testing, and code coverage requirements. |

## Key Architectural Principles

1. **V1 Scope Strictness:** V1 is dedicated solely to **historical transactional mirroring and continuous statement ingestion**. Payment initiation (ITP / Pix Initiation) is strictly out-of-scope for V1.
2. **Provider Decoupling:** The ingestion engine communicates with a unified `OpenFinanceProvider` interface, allowing instant switching between direct Open Finance and aggregators like Pluggy via the `PROVIDER` environment variable.
3. **Midaz Design Consistency:** Follows Midaz's Hexagonal/Clean Architecture, Fiber HTTP routing, OpenTelemetry observability, Zap structured logging, and CQRS service separation.
4. **Idempotent & Exact Double-Entry:** Every bank movement maps to a balanced transaction in Midaz anchored to an `@external` clearing account, with complete deduplication using banking transaction identifiers.
