# Accounting & Double-Entry Specification

## 1. Accounting Foundations in Midaz

Midaz is an immutable, multi-asset financial ledger built strictly upon the **double-entry bookkeeping** principles established by Luca Pacioli:
$$\sum \text{Debits} - \sum \text{Credits} = 0$$

Every economic event recorded in Midaz requires at least two counterbalancing entries. In standard accounts (e.g. user deposits, portfolios), Midaz rejects transactions that would drive balances below zero unless explicitly authorized.

---

## 2. Role of the `@external/` Account

Real-world bank accounts interact with the outside universe:
- Money enters via Pix, TED, or cash deposit from an external entity.
- Money leaves via Pix, bill payment (boleto), or transfer to an external bank.

In a closed ledger, money cannot materialize out of thin air or vanish into nothingness. Midaz solves this through **External Accounts** ([pkg/constant/account.go](file:///Users/adryanalencar/Documents/projetos%202/midaz/pkg/constant/account.go#L4)).

### Properties of `@external/` Accounts:
1. **Negative Balance Tolerance:** By design, Midaz permits accounts whose alias begins with `@external/` (or whose type is `external`) to hold negative balances.
2. **Clearance & Liquidity Mirroring:** The `@external` account acts as the mirror of the entire external financial universe.
3. **Dedicated Plugin Clearance:** To distinguish Open Finance synchronizations from internal ledger minting or manual operational adjustments, this plugin uses:
   $$\text{Alias: } \texttt{@external/BRL:openfinance}$$

---

## 3. Transaction Mapping Rules

### 3.1. Case A: Bank Inflow (`CREDITO` / Deposit)
* **Real-world event:** A customer transfers R$ 1,000.00 via Pix into the company's Itaú account.
* **Open Finance event:** `creditDebitType = "CREDITO"`, `amount = "1000.00"`.
* **Midaz Ledger entry:**
  - **Source (Debit):** `@external/BRL:openfinance` (Decreases / goes negative by R$ 1,000.00).
  - **Distribution (Credit):** `itau_main_account` (Increases by R$ 1,000.00).

```json
{
  "chartOfAccountsGroupName": "OPEN_FINANCE_SYNC",
  "description": "OFP Inflow: PIX RECEBIDO - CLIENTE ABC",
  "code": "OFP-TX-987123",
  "metadata": {
    "provider": "openfinance",
    "bank_transaction_id": "TX-987123",
    "payment_type": "PIX",
    "credit_debit_type": "CREDITO"
  },
  "send": {
    "asset": "BRL",
    "value": 100000,
    "scale": 2,
    "source": {
      "from": [
        {
          "accountAlias": "@external/BRL:openfinance",
          "amount": {
            "asset": "BRL",
            "value": "1000.00"
          }
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "itau_operating_account",
          "amount": {
            "asset": "BRL",
            "value": "1000.00"
          }
        }
      ]
    }
  }
}
```

---

### 3.2. Case B: Bank Outflow (`DEBITO` / Withdrawal / Payment)
* **Real-world event:** The company pays a supplier R$ 450.50 via Pix from its Itaú account.
* **Open Finance event:** `creditDebitType = "DEBITO"`, `amount = "450.50"`.
* **Midaz Ledger entry:**
  - **Source (Debit):** `itau_operating_account` (Decreases by R$ 450.50).
  - **Distribution (Credit):** `@external/BRL:openfinance` (Increases / pays back the clearing account by R$ 450.50).

```json
{
  "chartOfAccountsGroupName": "OPEN_FINANCE_SYNC",
  "description": "OFP Outflow: PIX ENVIADO - FORNECEDOR XYZ",
  "code": "OFP-TX-987124",
  "metadata": {
    "provider": "openfinance",
    "bank_transaction_id": "TX-987124",
    "payment_type": "PIX",
    "credit_debit_type": "DEBITO"
  },
  "send": {
    "asset": "BRL",
    "value": 45050,
    "scale": 2,
    "source": {
      "from": [
        {
          "accountAlias": "itau_operating_account",
          "amount": {
            "asset": "BRL",
            "value": "450.50"
          }
        }
      ]
    },
    "distribute": {
      "to": [
        {
          "accountAlias": "@external/BRL:openfinance",
          "amount": {
            "asset": "BRL",
            "value": "450.50"
          }
        }
      ]
    }
  }
}
```

---

### 3.3. Case C: Reversals / Chargebacks (`ESTORNO`)
When a transaction is marked as an `ESTORNO` by the banking institution, the plugin treats it as an inverse flow with an explicit reference in the metadata:
- A reversed debit becomes an inflow (`CREDITO` from `@external`).
- A reversed credit becomes an outflow (`DEBITO` to `@external`).
- The `metadata` field records `"is_reversal": true` and `"original_transaction_id": "<id>"`.

---

## 4. Fixed-Point Math & Currency Scale

Bank APIs return monetary values as floating decimal strings (e.g., `"150.25"` or `150.25`). 
Midaz operates with **fixed-point integer representation** coupled with a `scale` parameter to prevent IEEE 754 floating-point inaccuracies.

### Conversion Formula:
$$\text{integerValue} = \text{round}(\text{decimalAmount} \times 10^{\text{scale}})$$

For Brazilian Real (`BRL`), the standard scale is $2$:
- `"150.25"` BRL $\implies$ `value: 15025`, `scale: 2`
- `"10.00"` BRL $\implies$ `value: 1000`, `scale: 2`
- `"0.05"` BRL $\implies$ `value: 5`, `scale: 2`

The plugin internal parser strictly converts strings via Go's `shopspring/decimal` library (the exact same library used inside Midaz) before building the JSON payload.
