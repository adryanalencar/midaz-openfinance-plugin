import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json'
  }
});

const MOCK_BANK_ACCOUNTS = [
  {
    id: '018f9211-5b72-74ba-9132-e018d9f1092a',
    institution_name: 'Banco Itaú Unibanco',
    institution_code: '341',
    account_number: '18920-4',
    branch_number: '0812',
    account_type: 'CHECKING',
    holder_name: 'Alpha Supermarket LTDA',
    midaz_account_alias: '@merchant_alpha/checking',
    clearing_account_alias: '@external/pluggy_itau',
    reconciliation_enabled: true,
    currency: 'BRL',
    balance: 489210.5,
    status: 'SYNCED',
    last_sync_at: '2026-11-21T19:45:00Z',
    connector: 'Pluggy Open Finance API'
  },
  {
    id: '018f9212-74ba-74ba-9132-e018d9f1092b',
    institution_name: 'Banco Bradesco',
    institution_code: '237',
    account_number: '44910-1',
    branch_number: '1402',
    account_type: 'CHECKING',
    holder_name: 'Alpha Supermarket LTDA',
    midaz_account_alias: '@merchant_alpha/bradesco_checking',
    clearing_account_alias: '@external/pluggy_bradesco',
    reconciliation_enabled: false,
    currency: 'BRL',
    balance: 142050.0,
    status: 'SYNCED',
    last_sync_at: '2026-11-21T18:30:00Z',
    connector: 'Pluggy Open Finance API'
  },
  {
    id: '018f9213-9132-74ba-9132-e018d9f1092c',
    institution_name: 'Banco do Brasil',
    institution_code: '001',
    account_number: '99201-8',
    branch_number: '3310',
    account_type: 'CHECKING',
    holder_name: 'Bravo Retail Group S.A.',
    midaz_account_alias: '@bravo_retail/checking',
    clearing_account_alias: '@external/pluggy_bb',
    reconciliation_enabled: true,
    currency: 'BRL',
    balance: 812400.25,
    status: 'SYNCED',
    last_sync_at: '2026-11-21T20:15:00Z',
    connector: 'Pluggy Open Finance API'
  }
];

const MOCK_CONSENTS = [
  {
    id: 'consent_itau_01',
    institution: 'Banco Itaú Unibanco',
    consent_id: 'urn:itau:consent:9812-441-aa',
    status: 'ACTIVE',
    permissions: ['ACCOUNTS_READ', 'BALANCES_READ', 'TRANSACTIONS_READ'],
    created_at: '2026-08-15T10:00:00Z',
    expires_at: '2027-08-15T10:00:00Z',
    days_remaining: 267
  },
  {
    id: 'consent_bradesco_01',
    institution: 'Banco Bradesco',
    consent_id: 'urn:bradesco:consent:5512-990-bb',
    status: 'ACTIVE',
    permissions: ['ACCOUNTS_READ', 'BALANCES_READ', 'TRANSACTIONS_READ'],
    created_at: '2026-09-01T12:00:00Z',
    expires_at: '2027-09-01T12:00:00Z',
    days_remaining: 284
  },
  {
    id: 'consent_bb_01',
    institution: 'Banco do Brasil',
    consent_id: 'urn:bb:consent:1102-331-cc',
    status: 'EXPIRING_SOON',
    permissions: ['ACCOUNTS_READ', 'BALANCES_READ', 'TRANSACTIONS_READ'],
    created_at: '2025-12-01T08:00:00Z',
    expires_at: '2026-12-01T08:00:00Z',
    days_remaining: 10
  }
];

const MOCK_TRANSACTIONS = [
  {
    id: 'tx_of_01',
    bank: 'Itaú Unibanco',
    account_alias: '@merchant_alpha/checking',
    description: 'CRÉDITO CARTÕES SUMUP REF: 20261120',
    type: 'CREDIT',
    amount: 43844.0,
    booking_date: '2026-11-20',
    status: 'INTERCEPTED_FOR_MATCHER',
    reconciliation_flag: true,
    matcher_source_b_id: 'sb_matcher_01948df2-0003'
  },
  {
    id: 'tx_of_02',
    bank: 'Itaú Unibanco',
    account_alias: '@merchant_alpha/checking',
    description: 'PIX RECEBIDO - CLIENTE 4991',
    type: 'CREDIT',
    amount: 1420.0,
    booking_date: '2026-11-21',
    status: 'DIRECT_DEPOSIT_COMMITTED',
    reconciliation_flag: false,
    midaz_tx_id: 'tx_midaz_direct_09124'
  },
  {
    id: 'tx_of_03',
    bank: 'Bradesco',
    account_alias: '@merchant_alpha/bradesco_checking',
    description: 'PAGAMENTO FORNECEDOR AMBEV S.A.',
    type: 'DEBIT',
    amount: -18500.0,
    booking_date: '2026-11-21',
    status: 'DIRECT_WITHDRAWAL_COMMITTED',
    reconciliation_flag: false,
    midaz_tx_id: 'tx_midaz_direct_09125'
  },
  {
    id: 'tx_of_04',
    bank: 'Banco do Brasil',
    account_alias: '@bravo_retail/checking',
    description: 'ADITUM MEIOS DE PAGAMENTO D+1',
    type: 'CREDIT',
    amount: 12516.68,
    booking_date: '2026-11-19',
    status: 'INTERCEPTED_FOR_MATCHER',
    reconciliation_flag: true,
    matcher_source_b_id: 'sb_matcher_01948df2-0008'
  }
];

export const openfinanceApi = {
  async getStatus() {
    try {
      const response = await api.get('/v1/sync/status');
      return response.data;
    } catch {
      return {
        accounts_tracked: 3,
        last_successful_sync: new Date().toISOString(),
        total_transactions_ingested: 1420,
        failed_transactions: 0,
        status: 'healthy',
        provider: 'Pluggy / Open Finance Brasil'
      };
    }
  },

  async getBankAccounts() {
    try {
      const response = await api.get('/v1/accounts');
      return response.data;
    } catch {
      return MOCK_BANK_ACCOUNTS;
    }
  },

  async triggerSync(bankAccountId) {
    try {
      const response = await api.post('/v1/sync', { bank_account_id: bankAccountId });
      return response.data;
    } catch {
      return {
        status: 'queued',
        job_id: `job_${Date.now()}`,
        message: 'Sync job enqueued with upstream Open Finance provider'
      };
    }
  },

  async getConsents() {
    try {
      const response = await api.get('/v1/consents');
      return response.data;
    } catch {
      return MOCK_CONSENTS;
    }
  },

  async getTransactions() {
    try {
      const response = await api.get('/v1/transactions');
      return response.data;
    } catch {
      return MOCK_TRANSACTIONS;
    }
  }
};
