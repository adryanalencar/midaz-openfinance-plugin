import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import TextField from '@mui/material/TextField';
import InputAdornment from '@mui/material/InputAdornment';

// @project
import { BentoCard, StatusBadge } from '@/components/alignly';
import { openfinanceApi } from '@/api/openfinanceApi';

// @assets
import { IconSearch } from '@tabler/icons-react';

/***************************  TRANSACTIONS & MIRRORED STATEMENTS VIEW  ***************************/

export default function TransactionsView() {
  const [transactions, setTransactions] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    openfinanceApi.getTransactions().then(setTransactions);
  }, []);

  const filtered = transactions.filter((t) =>
    t.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
    t.bank.toLowerCase().includes(searchQuery.toLowerCase()) ||
    t.account_alias.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatCurrency = (val) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);

  return (
    <Box sx={{ pb: 6 }}>
      {/* Header */}
      <Stack
        direction={{ xs: 'column', sm: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'flex-start', sm: 'center' }}
        spacing={2}
        sx={{ mb: 4 }}
      >
        <Box>
          <Typography
            variant="h4"
            sx={{
              fontFamily: "'Plus Jakarta Sans', sans-serif",
              fontWeight: 700,
              color: '#1D1D1D',
              letterSpacing: '-0.03em',
              mb: 0.5
            }}
          >
            Bank Statement Transactions
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Ingested statement lines &amp; selective ledger routing audit
          </Typography>
        </Box>

        <TextField
          size="small"
          placeholder="Search statement..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <IconSearch size={16} color="#919EAB" />
              </InputAdornment>
            ),
            sx: { borderRadius: '9999px', bgcolor: '#FFFFFF', fontSize: '12px', width: 240 }
          }}
        />
      </Stack>

      {/* Main Table Bento Card */}
      <BentoCard variant="white" sx={{ p: 3 }}>
        <Stack spacing={1.5}>
          {filtered.map((tx) => (
            <Box
              key={tx.id}
              sx={{
                display: 'flex',
                flexDirection: { xs: 'column', md: 'row' },
                alignItems: { xs: 'flex-start', md: 'center' },
                justifyContent: 'space-between',
                p: 2,
                borderRadius: '16px',
                bgcolor: '#FFFFFF',
                border: '1px solid #E8E8E8',
                gap: 2,
                transition: 'all 0.15s ease',
                '&:hover': {
                  borderColor: '#D0D0D0',
                  boxShadow: '0 4px 16px rgba(0, 0, 0, 0.04)'
                }
              }}
            >
              <Stack direction="row" spacing={2} alignItems="center">
                <Box
                  sx={{
                    width: 44,
                    height: 44,
                    borderRadius: '12px',
                    bgcolor: tx.type === 'CREDIT' ? '#F0FD63' : '#1D1D1D',
                    color: tx.type === 'CREDIT' ? '#1D1D1D' : '#FFFFFF',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    fontWeight: 700,
                    fontSize: '12px',
                    flexShrink: 0
                  }}
                >
                  {tx.bank.slice(0, 3).toUpperCase()}
                </Box>
                <Box>
                  <Typography sx={{ fontWeight: 700, fontSize: '13px', color: '#1D1D1D' }}>
                    {tx.description}
                  </Typography>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB', mt: 0.25 }}>
                    {tx.bank} • {tx.account_alias} • {tx.booking_date}
                  </Typography>
                </Box>
              </Stack>

              <Box sx={{ textAlign: { xs: 'left', md: 'right' } }}>
                <Typography
                  sx={{
                    fontWeight: 700,
                    fontSize: '14px',
                    color: tx.amount > 0 ? '#15803d' : '#1D1D1D'
                  }}
                >
                  {formatCurrency(tx.amount)}
                </Typography>
                <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                  {tx.type}
                </Typography>
              </Box>

              <Box sx={{ alignSelf: { xs: 'flex-start', md: 'center' } }}>
                <StatusBadge variant={tx.reconciliation_flag ? 'lime' : 'neutral'}>
                  {tx.reconciliation_flag ? 'Routed to Matcher 6' : 'Direct Midaz Posting'}
                </StatusBadge>
              </Box>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
