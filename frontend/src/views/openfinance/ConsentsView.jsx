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
            Open Finance Consents
          </Typography>
          <Typography variant="body2" sx={{ color: '#919EAB' }}>
            Open Finance Brasil consent lifecycle &amp; automatic renewal triggers
          </Typography>
        </Box>

        <StatusBadge variant="lime" icon={<IconShieldCheck size={14} />}>
          Regulatory Compliant
        </StatusBadge>
      </Stack>

      <Grid container spacing={3}>
        {consents.map((consent) => (
          <Grid item xs={12} md={4} key={consent.id}>
            <BentoCard variant="white" sx={{ p: 3, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'space-between' }}>
              <Box>
                <Stack direction="row" justifyContent="space-between" alignItems="flex-start" sx={{ mb: 2 }}>
                  <Typography sx={{ fontFamily: "'Plus Jakarta Sans', sans-serif", fontWeight: 700, fontSize: '16px', color: '#1D1D1D' }}>
                    {consent.institution}
                  </Typography>
                  <StatusBadge variant={consent.status === 'ACTIVE' ? 'lime' : 'warning'}>
                    {consent.status === 'ACTIVE' ? 'Active Consent' : 'Expiring Soon'}
                  </StatusBadge>
                </Stack>

                <Typography sx={{ fontSize: '11px', color: '#919EAB', mb: 2 }}>
                  Consent ID: <code>{consent.consent_id}</code>
                </Typography>

                <Stack spacing={1} sx={{ p: 2, borderRadius: '14px', bgcolor: '#F3F3F3', mb: 2.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <IconClock size={15} color="#919EAB" />
                    <Typography sx={{ fontSize: '12px', color: '#666666' }}>
                      Expires In: <strong>{consent.days_remaining} days</strong>
                    </Typography>
                  </Stack>
                  <Typography sx={{ fontSize: '11px', color: '#919EAB' }}>
                    Valid until {new Date(consent.expires_at).toLocaleDateString()}
                  </Typography>

                  <Box sx={{ mt: 1 }}>
                    <Typography sx={{ fontSize: '11px', fontWeight: 600, color: '#1D1D1D', mb: 0.5 }}>
                      Authorized Scopes:
                    </Typography>
                    <Stack direction="row" spacing={0.5} flexWrap="wrap">
                      {consent.permissions.map((p) => (
                        <Box
                          key={p}
                          sx={{
                            px: 1,
                            py: 0.25,
                            borderRadius: '6px',
                            bgcolor: '#FFFFFF',
                            border: '1px solid #E8E8E8',
                            fontSize: '10px',
                            color: '#666666'
                          }}
                        >
                          {p}
                        </Box>
                      ))}
                    </Stack>
                  </Box>
                </Stack>
              </Box>

              <Button
                variant={consent.days_remaining < 30 ? 'contained' : 'outlined'}
                fullWidth
                startIcon={<IconRefresh size={14} />}
                sx={{
                  borderRadius: '9999px',
                  fontSize: '11px',
                  py: 1,
                  bgcolor: consent.days_remaining < 30 ? '#1D1D1D' : '#FFFFFF',
                  color: consent.days_remaining < 30 ? '#FFFFFF' : '#1D1D1D',
                  borderColor: '#E8E8E8',
                  '&:hover': {
                    bgcolor: consent.days_remaining < 30 ? '#000000' : '#F3F3F3'
                  }
                }}
              >
                {consent.days_remaining < 30 ? 'Renew Consent Now' : 'Check Consent Status'}
              </Button>
            </BentoCard>
          </Grid>
        ))}
      </Grid>
    </Box>
  );
}
