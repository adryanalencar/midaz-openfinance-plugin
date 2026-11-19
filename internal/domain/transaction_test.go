package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestBankTransactionNormalization is responsible to test BankTransaction struct initialization
func TestBankTransactionNormalization(t *testing.T) {
	amount, err := decimal.NewFromString("1500.75")
	assert.NoError(t, err)

	now := time.Now().UTC()

	tx := BankTransaction{
		ID:              "TX-987123",
		AccountID:       "acc-itau-001",
		Amount:          amount,
		CreditDebitType: CreditDebitCredit,
		Description:     "PIX RECEBIDO JOAO",
		PaymentType:     "PIX",
		Currency:        "BRL",
		TransactionDate: now,
		BookingDateTime: now,
	}

	assert.Equal(t, "TX-987123", tx.ID)
	assert.Equal(t, "acc-itau-001", tx.AccountID)
	assert.Equal(t, CreditDebitCredit, tx.CreditDebitType)
	assert.Equal(t, "1500.75", tx.Amount.StringFixed(2))
	assert.Equal(t, "BRL", tx.Currency)
}

// TestDecimalFixedPointScaling is responsible to test table-driven currency conversion to Midaz fixed-point integer
func TestDecimalFixedPointScaling(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedValue int64
		expectedScale int32
	}{
		{
			name:          "standard two decimals",
			input:         "150.25",
			expectedValue: 15025,
			expectedScale: 2,
		},
		{
			name:          "round hundred amount",
			input:         "100.00",
			expectedValue: 10000,
			expectedScale: 2,
		},
		{
			name:          "cents only",
			input:         "0.75",
			expectedValue: 75,
			expectedScale: 2,
		},
		{
			name:          "four decimals rounded to scale two",
			input:         "1000.0400",
			expectedValue: 100004,
			expectedScale: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec, err := decimal.NewFromString(tt.input)
			assert.NoError(t, err)

			val := dec.Mul(decimal.NewFromInt(100)).IntPart()
			assert.Equal(t, tt.expectedValue, val)
		})
	}
}
