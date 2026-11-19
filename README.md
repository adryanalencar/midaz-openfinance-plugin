# Midaz Plugin: Open Finance

An enterprise-grade plugin for the **Midaz Ledger** that synchronizes real-world banking statement lines into Midaz's double-entry accounting model via **Open Finance Brasil** (direct FAPI 1.0 mTLS) or **Pluggy** aggregator.

## Features
- **Strict V1 Scope:** Historical and continuous statement mirroring only (no payment initiation).
- **Double-Entry Guarantees:** Every deposit or withdrawal is balanced against `@external/BRL:openfinance`.
- **Pluggable Architecture:** Switch between `openfinance` (direct) and `pluggy` (aggregator) via the `PROVIDER` env var.
- **Midaz Console Dynamic Plugin:** Exposes `/api/manifest` for registration in the Midaz Console.
- **Midaz Code Standards:** Built with Go, Fiber, CQRS, Hexagonal Architecture, and OpenTelemetry tracing.

## Documentation
Comprehensive English documentation is located in the [`docs/`](docs/) directory:
- [01. Product Requirements Document](docs/01-PRODUCT-REQUIREMENTS.md)
- [02. Technical Architecture](docs/02-TECHNICAL-ARCHITECTURE.md)
- [03. Accounting & Double-Entry Specification](docs/03-ACCOUNTING-DOUBLE-ENTRY.md)
- [04. Open Finance Brasil Direct Specification](docs/04-OPEN-FINANCE-BRASIL-SPEC.md)
- [05. Multi-Provider Strategy & Extensibility](docs/05-PROVIDER-STRATEGY.md)
- [06. API & Plugin Manifest Specification](docs/06-API-AND-MANIFEST-SPEC.md)

## Quick Start

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Configure your provider (`PROVIDER=openfinance` or `PROVIDER=pluggy`) and Midaz connection IDs.

3. Run the application:
   ```bash
   go run cmd/app/main.go
   ```

4. Register with Midaz Console:
   ```bash
   curl -X POST http://localhost:3000/api/plugin/manifest/register \
     -H "Content-Type: application/json" \
     -d '{"host": "http://localhost:8085"}'
   ```
