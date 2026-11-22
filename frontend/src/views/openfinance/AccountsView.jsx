import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';

// @project
import { BentoCard, StatusBadge } from '@/components/alignly';
import { openfinanceApi } from '@/api/openfinanceApi';

// @assets
import {
  IconBuildingBank,
  IconRefresh,
  IconPlus,
  IconCheck,
  IconShieldCheck,
  IconWallet
} from '@tabler/icons-react';

/***************************  CONNECTED BANK ACCOUNTS VIEW  ***************************/

export default function AccountsView() {
  const [accounts, setAccounts] = useState([]);
  const [syncingId, setSyncingId] = useState(null);
  const [syncedFeedback, setSyncedFeedback] = useState(null);

  const loadAccounts = async () => {
    const data = await openfinanceApi.getBankAccounts();
    setAccounts(data || []);
  };

  useEffect(() => {
    loadAccounts();
  }, []);

  const handleSync = async (accId) => {
    setSyncingId(accId);
    try {
      await openfinanceApi.triggerSync(accId);
      setSyncedFeedback(accId);
      setTimeout(() => setSyncedFeedback(null), 3000);
      loadAccounts();
    } finally {
      setSyncingId(null);
    }
  };

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
            Contas Bancárias Conectadas
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Contas correntes integradas via Open Finance com atualização contínua de saldos e extratos
          </Typography>
        </Box>

        <Button
          variant="contained"
          startIcon={<IconPlus size={16} />}
          sx={{
            borderRadius: '9999px',
            bgcolor: '#1D1D1D',
            color: '#FFFFFF',
            px: 3,
            '&:hover': { bgcolor: '#000000' }
          }}
        >
          Conectar Nova Conta Bancária
        </Button>
      </Stack>

      {/* Grid of Bank Accounts (3 cards per row on desktop/laptop) */}
      <Grid container spacing={3}>
        {accounts.map((acc) => {
          const isSyncing = syncingId === acc.id;
          const isRecentlySynced = syncedFeedback === acc.id;

          return (
            <Grid size={{ xs: 12, sm: 6, lg: 4 }} key={acc.id}>
              <BentoCard
                variant="white"
                sx={{
                  p: 3.5,
                  height: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'space-between'
                }}
              >
                <Box>
                  <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                    <Stack direction="row" spacing={1.5} alignItems="center">
                      <Box
                        sx={{
                          width: 44,
                          height: 44,
                          borderRadius: '12px',
                          bgcolor: '#1D1D1D',
                          color: '#F0FD63',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 700,
                          fontSize: '13px'
                        }}
                      >
                        <IconBuildingBank size={22} />
                      </Box>
                      <Box>
                        <Typography
                          sx={{
                            fontFamily: "'Plus Jakarta Sans', sans-serif",
                            fontWeight: 700,
                            fontSize: '16px',
                            color: '#1D1D1D'
                          }}
                        >
                          {acc.institution_name}
                        </Typography>
                        <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                          Agência: {acc.branch_number} • Conta: {acc.account_number}
                        </Typography>
                      </Box>
                    </Stack>
                    <StatusBadge variant="lime">
                      Conectado
                    </StatusBadge>
                  </Stack>

                  {/* Account Information Details */}
                  <Box sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                    <Stack spacing={1.2}>
                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <IconWallet size={15} color="#666666" />
                          <Typography sx={{ fontSize: '12px', color: '#666666' }}>Modalidade:</Typography>
                        </Stack>
                        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#1D1D1D' }}>
                          Conta Corrente PJ
                        </Typography>
                      </Box>

                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <IconShieldCheck size={15} color="#666666" />
                          <Typography sx={{ fontSize: '12px', color: '#666666' }}>Conciliação de Cartão:</Typography>
                        </Stack>
                        <StatusBadge variant={acc.reconciliation_enabled ? 'lime' : 'neutral'}>
                          {acc.reconciliation_enabled ? 'Automática Ativa' : 'Depósito Direto'}
                        </StatusBadge>
                      </Box>

                      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <IconBuildingBank size={15} color="#666666" />
                          <Typography sx={{ fontSize: '12px', color: '#666666' }}>Conexão:</Typography>
                        </Stack>
                        <Typography sx={{ fontSize: '12px', fontWeight: 600, color: '#15803d' }}>
                          Open Finance Brasil
                        </Typography>
                      </Box>
                    </Stack>
                  </Box>
                </Box>

                <Stack
                  direction="row"
                  justifyContent="space-between"
                  alignItems="center"
                  sx={{ pt: 2, borderTop: '1px solid #E8E8E8' }}
                >
                  <Box>
                    <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>Saldo Disponível</Typography>
                    <Typography sx={{ fontSize: '18px', fontWeight: 700, color: '#1D1D1D' }}>
                      {formatCurrency(acc.balance)}
                    </Typography>
                  </Box>

                  <Button
                    size="small"
                    variant="outlined"
                    onClick={() => handleSync(acc.id)}
                    disabled={isSyncing}
                    startIcon={
                      isSyncing ? (
                        <CircularProgress size={14} color="inherit" />
                      ) : isRecentlySynced ? (
                        <IconCheck size={14} color="#15803d" />
                      ) : (
                        <IconRefresh size={14} />
                      )
                    }
                    sx={{
                      borderRadius: '9999px',
                      color: isRecentlySynced ? '#15803d' : '#1D1D1D',
                      borderColor: isRecentlySynced ? 'rgba(34, 197, 94, 0.4)' : '#E8E8E8',
                      bgcolor: isRecentlySynced ? 'rgba(34, 197, 94, 0.08)' : '#FFFFFF',
                      fontSize: '11px',
                      py: 0.5,
                      px: 2,
                      '&:hover': { bgcolor: '#F3F3F3' }
                    }}
                  >
                    {isSyncing ? 'Atualizando...' : isRecentlySynced ? 'Atualizado!' : 'Atualizar Extrato'}
                  </Button>
                </Stack>
              </BentoCard>
            </Grid>
          );
        })}
      </Grid>
    </Box>
  );
}
