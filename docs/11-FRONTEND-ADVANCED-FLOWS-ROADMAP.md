# Roadmap: Advanced Frontend Flows for Midaz Open Finance Plugin

This document outlines the advanced operational flows and edge cases scheduled for future implementation in `frontend/`.

---

## 1. Advanced Sync Trigger with Date Range Filters
- **Objective:** Allow operators to trigger targeted statement synchronizations across arbitrary historical dates instead of standard polling windows.
- **Components to Add:**
  - `ManualSyncRangeDialog`:
    - Select target institution (e.g., Itaú, Bradesco, BB).
    - Date Pickers for `from_date` and `to_date`.
    - Dispatches payload `{ bank_account_id, from_date, to_date }` to `POST /v1/sync`.
    - Real-time polling of queued job status.

---

## 2. Dynamic `reconciliation.enabled` Routing Switch
- **Objective:** Enable toggling whether an individual bank account routes acquirer settlement deposits to Matcher 6 or executes direct Ledger deposits.
- **Components to Add:**
  - Interactive switch toggle on the Bank Account card.
  - Confirmation dialog explaining the double-credit implications.
  - Updates account metadata flag in Midaz and local plugin repository.

---

## 3. Asynchronous Sync Jobs & Error Audit Queue
- **Objective:** Visibility into background Open Finance synchronization workers and upstream banking API availability.
- **Components to Add:**
  - `SyncJobsAuditView`:
    - Table displaying `job_id`, `bank_account_id`, `status` (`QUEUED`, `RUNNING`, `COMPLETED`, `FAILED`).
    - Error details inspector (e.g. invalid credentials, MFA required, bank server timeout).
    - Retry job trigger button.

---

## 4. Pluggy Connect Widget Native Integration
- **Objective:** Embed the native Pluggy Connect SDK widget inside the frontend modal for frictionless bank account linking.
- **Components to Add:**
  - `PluggyConnectModal`: Loads `https://cdn.pluggy.ai/pluggy-connect.js`.
  - Handles `onSuccess` callback to automatically register new credentials and retrieve consent ID.
