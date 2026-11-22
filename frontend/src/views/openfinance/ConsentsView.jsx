import { useState, useEffect } from 'react';

// @mui
import Box from '@mui/material/Box';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';

// @project
import { BentoCard, StatusBadge } from '@/components/alignly';
import { openfinanceApi } from '@/api/openfinanceApi';

// @assets
import {
  IconShieldCheck,
  IconClock,
  IconCheck,
  IconRefresh
} from '@tabler/icons-react';

/***************************  OPEN FINANCE CONSENTS VIEW  ***************************/

export default function ConsentsView() {
  const [consents, setConsents] = useState([]);

  useEffect(() => {
    openfinanceApi.getConsents().then(setConsents);
  }, []);

  const permissionLabels = {
    ACCOUNTS_READ: 'Consulta de Dados da Conta',
    BALANCES_READ: 'Saldo Consolidado em Tempo Real',
    TRANSACTIONS_READ: 'Extrato e Movimentações Financeiras'
  };

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
            Autorizações Open Finance
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Consentimentos bancários regulamentados pelo Banco Central para leitura de extratos
          </Typography>
        </Box>

        <StatusBadge variant="lime" icon={<IconShieldCheck size={14} />}>
          Em Conformidade com Banco Central
        </StatusBadge>
      </Stack>

      <Grid container spacing={3}>
        {consents.map((consent) => (
          <Grid size={{ xs: 12, sm: 6, lg: 4 }} key={consent.id}>
            <BentoCard
              variant="white"
              sx={{
                p: 3,
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'space-between'
              }}
            >
              <Box>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                  <Typography
                    sx={{
                      fontFamily: "'Plus Jakarta Sans', sans-serif",
                      fontWeight: 700,
                      fontSize: '16px',
                      color: '#1D1D1D'
                    }}
                  >
                    {consent.institution}
                  </Typography>
                  <StatusBadge variant={consent.status === 'ACTIVE' ? 'lime' : 'warning'}>
                    {consent.status === 'ACTIVE' ? 'Autorização Ativa' : 'Renovação Próxima'}
                  </StatusBadge>
                </Stack>

                <Stack spacing={1} sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <IconClock size={15} color="#919EAB" />
                    <Typography sx={{ fontSize: '12px', color: '#666666' }}>
                      Expira em: <strong>{consent.days_remaining} dias</strong>
                    </Typography>
                  </Stack>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                    Válido até {new Date(consent.expires_at).toLocaleDateString('pt-BR')}
                  </Typography>

                  <Box sx={{ mt: 1 }}>
                    <Typography sx={{ fontSize: '11px', fontWeight: 600, color: '#1D1D1D', mb: 0.75 }}>
                      Permissões Ativas:
                    </Typography>
                    <Stack spacing={0.5}>
                      {consent.permissions.map((p) => (
                        <Stack key={p} direction="row" spacing={1} alignItems="center">
                          <IconCheck size={13} color="#15803d" />
                          <Typography sx={{ fontSize: '11px', color: '#666666' }}>
                            {permissionLabels[p] || p}
                          </Typography>
                        </Stack>
                      ))}
                    </Stack>
                  </Box>
                </Stack>
              </Box>

              <Button
                variant={consent.status === 'ACTIVE' ? 'outlined' : 'contained'}
                fullWidth
                startIcon={<IconRefresh size={14} />}
                sx={{
                  borderRadius: '9999px',
                  ...(consent.status === 'ACTIVE'
                    ? {
                        borderColor: '#E8E8E8',
                        color: '#1D1D1D',
                        bgcolor: '#FFFFFF',
                        '&:hover': { bgcolor: '#F3F3F3' }
                      }
                    : {
                        bgcolor: '#1D1D1D',
                        color: '#FFFFFF',
                        '&:hover': { bgcolor: '#000000' }
                      }),
                  fontSize: '12px',
                  py: 0.75
                }}
              >
                {consent.status === 'ACTIVE' ? 'Verificar Status da Conexão' : 'Renovar Autorização'}
              </Button>
            </BentoCard>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
