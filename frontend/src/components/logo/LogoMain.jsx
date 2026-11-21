// @mui
import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import Stack from '@mui/material/Stack';

/***************************  LOGO - MAIN (OPEN FINANCE)  ***************************/

export default function LogoMain() {
  return (
    <Stack direction="row" spacing={1.25} alignItems="center" sx={{ textDecoration: 'none' }}>
      <Box
        sx={{
          width: 32,
          height: 32,
          borderRadius: '9px',
          bgcolor: '#F0FD63',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          flexShrink: 0
        }}
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
          <path
            d="M3 21h18M3 10h18M5 10v11M19 10v11M9 10v11M14 10v11M4 10l8-7 8 7"
            stroke="#1D1D1D"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
      </Box>
      <Stack spacing={0}>
        <Typography
          sx={{
            color: '#FFFFFF',
            fontFamily: "'Plus Jakarta Sans', sans-serif",
            fontWeight: 700,
            fontSize: '15px',
            letterSpacing: '-0.02em',
            lineHeight: 1.1
          }}
        >
          Midaz
        </Typography>
        <Typography
          sx={{
            color: '#F0FD63',
            fontFamily: "'Plus Jakarta Sans', sans-serif",
            fontWeight: 600,
            fontSize: '10px',
            textTransform: 'uppercase',
            letterSpacing: '0.08em',
            lineHeight: 1
          }}
        >
          Open Finance
        </Typography>
      </Stack>
    </Stack>
  );
}
