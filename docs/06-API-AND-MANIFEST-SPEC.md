# API & Plugin Manifest Specification

## 1. Midaz Console Manifest Specification

The plugin serves `GET /api/manifest` so that the Midaz Console can dynamically discover and register it without requiring rebuilding or redeployment of Midaz.

The payload matches [PluginManifestEntity](file:///Users/adryanalencar/Documents/projetos%202/midaz/components/console/src/core/domain/entities/plugin-manifest-entity.ts):

### Manifest Endpoint: `GET /api/manifest`
* **Response Status:** `200 OK`
* **Response Body:**
```json
{
  "name": "plugin-openfinance",
  "title": "Open Finance Sync",
  "description": "Automated bank statement mirroring and double-entry ledger reconciliation",
  "version": "1.0.0",
  "route": "/openfinance",
  "entry": "http://localhost:8085/remoteEntry.js",
  "healthcheck": "/health",
  "host": "http://localhost:8085",
  "icon": "Landmark",
  "enabled": true,
  "author": "Midaz Community"
}
```

---

## 2. Dynamic Registration in Midaz Console

Once the plugin service is running, invoke Midaz Console's registration API:

```bash
curl -X POST http://localhost:3000/api/plugin/manifest/register \
  -H "Content-Type: application/json" \
  -d '{
    "host": "http://plugin-openfinance:8085"
  }'
```

The Midaz Console will:
1. Query `http://plugin-openfinance:8085/api/manifest`.
2. Persist the manifest into its internal MongoDB plugin collection.
3. Automatically inject the **Open Finance Sync** item into the sidebar menu with the `Landmark` icon.

---

## 3. Operational HTTP Endpoints

### 3.1. Health Check: `GET /health`
Validates plugin status, database connectivity, and upstream provider readiness.

* **Response Status:** `200 OK`
* **Response Body:**
```json
{
  "status": "healthy",
  "provider": "openfinance",
  "uptime": "14h22m10s",
  "midaz_connectivity": true,
  "clearing_account_verified": true
}
```

---

### 3.2. Manual Sync Trigger: `POST /v1/sync`
Allows triggering an immediate synchronization run for a specific bank account or all registered accounts.

* **Request Body:**
```json
{
  "bank_account_id": "018f9211-5b72-74ba-9132-e018d9f1092a",
  "from_date": "2026-11-01",
  "to_date": "2026-11-19"
}
```
* **Response Status:** `202 Accepted`
* **Response Body:**
```json
{
  "status": "queued",
  "job_id": "job_01jnd04918fskamv"
}
```

---

### 3.3. Sync Status & Audit: `GET /v1/sync/status`
Returns recent synchronization history and statistics.

* **Response Status:** `200 OK`
* **Response Body:**
```json
{
  "accounts_tracked": 3,
  "last_successful_sync": "2026-11-19T03:30:00Z",
  "total_transactions_ingested": 1420,
  "failed_transactions": 0,
  "recent_batches": [
    {
      "bank_account_id": "itau-cc-001",
      "records_fetched": 12,
      "records_posted": 12,
      "records_skipped": 0,
      "synced_at": "2026-11-19T03:30:00Z"
    }
  ]
}
```
