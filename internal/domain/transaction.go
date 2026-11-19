package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// CreditDebitType represents the direction of the transaction from the bank perspective.
type CreditDebitType string

const (
	CreditDebitCredit CreditDebitType = "CREDITO"
	CreditDebitDebit  CreditDebitType = "DEBITO"
)

// BankTransaction represents a normalized banking transaction received from any provider.
type BankTransaction struct {
	ID               string          `json:"id"`
	AccountID        string          `json:"accountId"`
	Amount           decimal.Decimal `json:"amount"`
	CreditDebitType  CreditDebitType `json:"creditDebitType"`
	Description      string          `json:"description"`
	PaymentType      string          `json:"paymentType"`
	Currency         string          `json:"currency"`
	TransactionDate  time.Time       `json:"transactionDate"`
	BookingDateTime  time.Time       `json:"bookingDateTime"`
	RawMetadata      map[string]any  `json:"rawMetadata,omitempty"`
}

// TransactionPage represents a paginated list of bank transactions.
type TransactionPage struct {
	Transactions []BankTransaction `json:"transactions"`
	CurrentPage  int               `json:"currentPage"`
	TotalPages   int               `json:"totalPages"`
	TotalRecords int               `json:"totalRecords"`
}

// BankAccount represents an external bank account discovered through Open Finance.
type BankAccount struct {
	ID          string `json:"id"`
	BrandName   string `json:"brandName"`
	CompeCode   string `json:"compeCode"`
	BranchCode  string `json:"branchCode"`
	Number      string `json:"number"`
	CheckDigit  string `json:"checkDigit"`
	Type        string `json:"type"`
	Currency    string `json:"currency"`
}
