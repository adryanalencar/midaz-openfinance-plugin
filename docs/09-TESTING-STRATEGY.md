# Testing Strategy & Quality Assurance Standards

Quality and correctness are non-negotiable within financial ledger systems. A single floating-point inaccuracy or unhandled nil pointer can corrupt account balances.

The **Midaz Open Finance Plugin** mandates comprehensive test coverage following the exact methodologies and tooling established across Midaz Core (`github.com/LerianStudio/midaz/v3`).

---

## 1. The Testing Pyramid

```
                / \
               /   \
              / E2E \           • Full Flow: Bank Sandbox -> Plugin -> Midaz Local
             /-------\
            / Integr. \         • HTTP API integration (httptest)
           /-----------\        • Midaz REST Client & Double-Entry Invariants
          /  Unit Tests \       • Domain Logic, Fixed-Point Decimal Math,
         /---------------\        Factory Pattern, JWS/mTLS Parsers
```

---

## 2. Test Tooling & Libraries

In strict alignment with [Midaz Core go.mod](file:///Users/adryanalencar/Documents/projetos%202/midaz/go.mod), all tests use:

| Library | Role | Usage in Plugin |
| :--- | :--- | :--- |
| **`testing`** | Standard Go Test Runner | Core runner (`go test -v -race ./...`) |
| **`github.com/stretchr/testify/assert`** | Assertions | Value checks (`assert.Equal`, `assert.NoError`, `assert.Nil`) |
| **`github.com/stretchr/testify/require`** | Hard Assertions | Fatal checks that abort test execution on unexpected nil |
| **`go.uber.org/mock/gomock`** | Mock Framework | Mocking repository ports and `OpenFinanceProvider` |
| **`net/http/httptest`** | HTTP Server Mocking | Mocking Midaz Onboarding/Transaction APIs and Open Finance endpoints |

---

## 3. Testing Standards & Conventions

### 3.1. Test Function Naming & Documentation
Every test function must have a descriptive comment explaining what is being verified, adhering to the Midaz standard:

```go
// TestCreateTransactionSuccess is responsible to test CreateTransaction with success
func TestCreateTransactionSuccess(t *testing.T) { ... }

// TestCreateTransactionError is responsible to test CreateTransaction with error
func TestCreateTransactionError(t *testing.T) { ... }
```

### 3.2. Table-Driven Tests Pattern
For complex mappings (such as converting bank amounts to Midaz integer scale, or classifying credit vs. debit), engineers must use table-driven tests:

```go
func TestDecimalConversion(t *testing.T) {
	tests := []struct {
		name          string
		inputAmount   string
		expectedValue int64
		expectedScale int32
	}{
		{
			name:          "standard two decimals",
			inputAmount:   "150.25",
			expectedValue: 15025,
			expectedScale: 2,
		},
		{
			name:          "integer round amount",
			inputAmount:   "100.00",
			expectedValue: 10000,
			expectedScale: 2,
		},
		{
			name:          "sub-cent four decimals truncated or scaled",
			inputAmount:   "0.05",
			expectedValue: 5,
			expectedScale: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// verification logic
		})
	}
}
```

---

## 4. Double-Entry Accounting Invariant Tests

All tests asserting transaction creation in Midaz must explicitly verify the **Luca Pacioli Invariant**:

$$\sum \text{Source Debits} = \sum \text{Distribute Credits}$$

```go
func AssertDoubleEntryBalanced(t *testing.T, payload map[string]any) {
	send := payload["send"].(map[string]any)
	totalValue := send["value"].(int64)

	source := send["source"].(map[string]any)
	fromList := source["from"].([]map[string]any)

	distribute := send["distribute"].(map[string]any)
	toList := distribute["to"].([]map[string]any)

	assert.NotEmpty(t, fromList, "Source accounts list must not be empty")
	assert.NotEmpty(t, toList, "Destination accounts list must not be empty")

	// Ensure clearing account is balanced against target
	hasExternalClearing := false
	for _, from := range fromList {
		if strings.HasPrefix(from["accountAlias"].(string), "@external/") {
			hasExternalClearing = true
		}
	}
	for _, to := range toList {
		if strings.HasPrefix(to["accountAlias"].(string), "@external/") {
			hasExternalClearing = true
		}
	}
	assert.True(t, hasExternalClearing, "Every external bank transaction must involve an @external account")
}
```

---

## 5. Coverage Requirements

* **Domain & Business Logic (`internal/domain/`, `internal/services/`):** Minimum **90% coverage**.
* **Adapters & Providers (`internal/adapters/`):** Minimum **80% coverage**.
* **Total Project Coverage Threshold:** **85%**.
* CI builds automatically fail if total test coverage drops below the 80% threshold.
