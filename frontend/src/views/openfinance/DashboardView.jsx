import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';

// @project
import { BentoCard, KpiCard, StatusBadge, PillTabs } from '@/components/alignly';
import { openfinanceApi } from '@/api/openfinanceApi';

// @assets
import {
  IconBuildingBank,
  IconRefresh,
  IconArrowsExchange,
  IconCheck,
  IconShieldCheck,
  IconArrowUpRight
} from '@tabler/icons-react';

/***************************  OPEN FINANCE DASHBOARD VIEW  ***************************/

export default function DashboardView() {
  const [loading, setLoading] = useState(false);
  const [syncing, setSyncing] = useState(false);
  const [statusData, setStatusData] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [activeTab, setActiveTab] = useState('all');

  const tabs = [
    { label: 'Todos os Lançamentos', value: 'all' },
    { label: 'Vendas no Cartão', value: 'matcher' },
    { label: 'Depósitos e PIX', value: 'direct' }
  ];

  const loadData = async () => {
    setLoading(true);
    try {
      const [statusRes, txRes] = await Promise.all([
        openfinanceApi.getStatus(),
        openfinanceApi.getTransactions()
      ]);
      setStatusData(statusRes);
      setTransactions(txRes);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const handleManualSync = async () => {
    setSyncing(true);
    try {
      await openfinanceApi.triggerSync();
      await loadData();
    } finally {
      setSyncing(false);
    }
  };

  const filteredTransactions = transactions.filter((tx) => {
    if (activeTab === 'matcher') return tx.reconciliation_flag;
    if (activeTab === 'direct') return !tx.reconciliation_flag;
    return true;
  });

  const formatCurrency = (val) =>
    new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(val);

  return (
    <Box sx={{ pb: 6 }}>
      {/* Header */}
      <Stack
        direction={{ xs: 'column', md: 'row' }}
        justifyContent="space-between"
        alignItems={{ xs: 'flex-start', md: 'center' }}
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
            Gestão Open Finance e Extratos
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Acompanhamento consolidado de saldos bancários e conciliação automática com maquininhas
          </Typography>
        </Box>

        <Stack direction="row" spacing={1.5} alignItems="center" flexWrap="wrap" gap={1}>
          <Button
            variant="contained"
            onClick={handleManualSync}
            disabled={syncing}
            startIcon={syncing ? <CircularProgress size={16} color="inherit" /> : <IconRefresh size={16} />}
            sx={{
              borderRadius: '9999px',
              bgcolor: '#1D1D1D',
              color: '#FFFFFF',
              px: 3,
              '&:hover': { bgcolor: '#000000' }
            }}
          >
            {syncing ? 'Sincronizando extratos...' : 'Sincronizar Bancos'}
          </Button>
          <PillTabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab} />
        </Stack>
      </Stack>

      {/* KPI Bento Grid */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Bancos Conectados"
            value="3"
            subtitle="Itaú, Bradesco e Banco do Brasil"
            badgeText="Sincronizados"
            badgeVariant="lime"
            variant="white"
            icon={<IconBuildingBank size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Lançamentos no Mês"
            value="1.420"
            subtitle="Extratos 100% atualizados"
            badgeText="Atualizado"
            badgeVariant="success"
            variant="white"
            icon={<IconCheck size={20} />}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <KpiCard
            title="Saldo Bancário Consolidado"
            value="R$ 1.443.660,75"
            subtitle="Disponível em conta corrente"
            badgeText="Tempo Real"
            badgeVariant="dark"
            variant="lime"
            icon={<IconArrowsExchange size={20} />}
          />
        </Grid>
      </Grid>

      {/* Featured Bento Highlight (Reconciliation Antiduplication) */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid size={{ xs: 12, lg: 7 }}>
          <BentoCard variant="dark" sx={{ height: '100%', p: 3.5 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
              <Box>
                <StatusBadge variant="lime" sx={{ mb: 1.5 }}>
                  Proteção Antiduplicação Ativa
                </StatusBadge>
                <Typography
                  sx={{
                    fontFamily: "'Plus Jakarta Sans', sans-serif",
                    fontSize: '20px',
                    fontWeight: 700,
                    color: '#FFFFFF',
                    letterSpacing: '-0.02em',
                    mb: 1
                  }}
                >
                  Conciliação Inteligente entre Maquininhas e Bancos
                </Typography>
                <Typography sx={{ color: '#919EAB', fontSize: '13px', lineHeight: 1.6, maxWidth: 540 }}>
                  O Midaz cruza automaticamente os comprovantes de venda das suas maquininhas de cartão com os depósitos bancários. Assim, seu caixa e seu extrato ficam sempre perfeitamente conciliados, eliminando duplicidade de saldo.
                </Typography>
              </Box>
              <Box
                sx={{
                  p: 1.5,
                  borderRadius: '16px',
                  bgcolor: '#262626',
                  color: '#F0FD63',
                  display: { xs: 'none', sm: 'flex' }
                }}
              >
                <IconShieldCheck size={28} />
              </Box>
            </Stack>

            <Stack direction="row" spacing={3} sx={{ pt: 2, borderTop: '1px solid rgba(255, 255, 255, 0.1)' }}>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase' }}>Origem das Vendas</Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  Maquininhas Integradas
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase' }}>Destino dos Depósitos</Typography>
                <Typography sx={{ color: '#FFFFFF', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  Contas Bancárias Verificadas
                </Typography>
              </Box>
              <Box>
                <Typography sx={{ color: '#919EAB', fontSize: '11px', textTransform: 'uppercase' }}>Status da Proteção</Typography>
                <Typography sx={{ color: '#F0FD63', fontWeight: 600, fontSize: '14px', mt: 0.5 }}>
                  100% Automática
                </Typography>
              </Box>
            </Stack>
          </BentoCard>
        </Grid>

        <Grid size={{ xs: 12, lg: 5 }}>
          <BentoCard variant="white" sx={{ height: '100%', p: 3.5, display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
            <Box>
              <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                <Typography sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '16px' }}>
                  Autorizações Open Finance
                </Typography>
                <StatusBadge variant="lime">
                  3 Ativas
                </StatusBadge>
              </Stack>

              <Stack spacing={2}>
                {[
                  { name: 'Banco Itaú', type: 'Conta Corrente & Extrato', days: '267 dias restantes', variant: 'lime' },
                  { name: 'Banco Bradesco', type: 'Conta Corrente & Extrato', days: '284 dias restantes', variant: 'lime' },
                  { name: 'Banco do Brasil', type: 'Conta Corrente & Extrato', days: '10 dias restantes', variant: 'pending' }
                ].map((c) => (
                  <Box
                    key={c.name}
                    sx={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      p: 1.5,
                      borderRadius: '14px',
                      bgcolor: '#F3F3F3'
                    }}
                  >
                    <Stack direction="row" spacing={1.5} alignItems="center">
                      <Box
                        sx={{
                          width: 32,
                          height: 32,
                          borderRadius: '10px',
                          bgcolor: '#1D1D1D',
                          color: '#FFFFFF',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 700,
                          fontSize: '11px'
                        }}
                      >
                        {c.name.split(' ')[1]?.[0] || 'B'}
                      </Box>
                      <Box>
                        <Typography sx={{ fontSize: '13px', fontWeight: 600, color: '#1D1D1D' }}>
                          {c.name}
                        </Typography>
                        <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                          {c.type}
                        </Typography>
                      </Box>
                    </Stack>
                    <StatusBadge variant={c.variant}>
                      {c.days}
                    </StatusBadge>
                  </Box>
                ))}
              </Stack>
            </Box>

            <Button
              variant="text"
              href="/openfinance/consents"
              endIcon={<IconArrowUpRight size={16} />}
              sx={{
                alignSelf: 'flex-start',
                mt: 2,
                p: 0,
                color: '#1D1D1D',
                fontWeight: 600,
                fontSize: '12px',
                '&:hover': { bgcolor: 'transparent', textDecoration: 'underline' }
              }}
            >
              Gerenciar Autorizações e Renovações
            </Button>
          </BentoCard>
        </Grid>
      </Grid>

      {/* Transactions & Statement Mirror Table */}
      <BentoCard variant="white" sx={{ p: 3 }}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 3 }}>
          <Box>
            <Typography
              variant="h6"
              sx={{
                fontFamily: "'Plus Jakarta Sans', sans-serif",
                fontWeight: 700,
                color: '#1D1D1D',
                letterSpacing: '-0.02em'
              }}
            >
              Últimas Movimentações Bancárias
            </Typography>
            <Typography variant="caption" sx={{ color: '#919EAB' }}>
              Extrato bancário em tempo real conciliado com o Midaz
            </Typography>
          </Box>
          <StatusBadge variant="lime">
            Tempo Real
          </StatusBadge>
        </Stack>

        <Stack spacing={1.5}>
          {filteredTransactions.map((tx) => (
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
                    {tx.bank} • {tx.booking_date}
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
                  {tx.type === 'CREDIT' ? 'Crédito em Conta' : 'Débito'}
                </Typography>
              </Box>

              <Box sx={{ alignSelf: { xs: 'flex-start', md: 'center' } }}>
                <StatusBadge variant={tx.reconciliation_flag ? 'lime' : 'neutral'}>
                  {tx.reconciliation_flag ? 'Venda de Cartão Conciliada' : 'Depósito em Conta'}
                </StatusBadge>
              </Box>
            </Box>
          ))}
        </Stack>
      </BentoCard>
    </Box>
  );
}
