package statement

// AccountListResponse matches ResponseAccountList in accounts-2.4.2.yml
type AccountListResponse struct {
	Data  []AccountData     `json:"data"`
	Links map[string]string `json:"links,omitempty"`
	Meta  PaginationMeta    `json:"meta,omitempty"`
}

type AccountData struct {
	BrandName   string `json:"brandName"`
	CompanyCnpj string `json:"companyCnpj"`
	Type        string `json:"type"`
	CompeCode   string `json:"compeCode"`
	BranchCode  string `json:"branchCode"`
	Number      string `json:"number"`
	CheckDigit  string `json:"checkDigit"`
	AccountID   string `json:"accountId"`
}

// AccountBalancesResponse matches ResponseAccountBalances in accounts-2.4.2.yml
type AccountBalancesResponse struct {
	Data struct {
		AvailableAmount           string `json:"availableAmount"`
		AvailableAmountCurrency   string `json:"availableAmountCurrency"`
		BlockedAmount             string `json:"blockedAmount"`
		AutomaticallyInvested     string `json:"automaticallyInvestedAmount"`
	} `json:"data"`
}

// AccountTransactionsResponse matches ResponseAccountTransactions in accounts-2.4.2.yml
type AccountTransactionsResponse struct {
	Data  []AccountTransactionData `json:"data"`
	Links struct {
		Self string `json:"self,omitempty"`
		Next string `json:"next,omitempty"`
		Prev string `json:"prev,omitempty"`
	} `json:"links,omitempty"`
	Meta PaginationMeta `json:"meta,omitempty"`
}

type AccountTransactionData struct {
	TransactionID                  string `json:"transactionId"`
	CompletedAuthorisedPaymentType string `json:"completedAuthorisedPaymentType"`
	CreditDebitType                string `json:"creditDebitType"` // CREDITO or DEBITO
	TransactionName                string `json:"transactionName"`
	Type                           string `json:"type"`
	TransactionAmount              struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"transactionAmount"`
	TransactionDateTime string `json:"transactionDateTime"`
	PartieCnpjCpf       string `json:"partieCnpjCpf,omitempty"`
	PartiePersonType    string `json:"partiePersonType,omitempty"`
	PartieCompeCode     string `json:"partieCompeCode,omitempty"`
	PartieBranchCode    string `json:"partieBranchCode,omitempty"`
	PartieNumber        string `json:"partieNumber,omitempty"`
	PartieCheckDigit    string `json:"partieCheckDigit,omitempty"`
}

type PaginationMeta struct {
	TotalRecords int `json:"totalRecords,omitempty"`
	TotalPages   int `json:"totalPages,omitempty"`
}
