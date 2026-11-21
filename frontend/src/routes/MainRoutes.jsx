import { lazy } from 'react';
import { Navigate } from 'react-router-dom';

// @project
import Loadable from '@/components/Loadable';
import { AuthProvider } from '@/contexts/AuthContext';
import AdminLayout from '@/layouts/AdminLayout';
import AuthGuard from '@/utils/route-guard/AuthGuard';
import RoleGuard from '@/utils/route-guard/RoleGuard';

// Views
const OpenFinanceDashboard = Loadable(lazy(() => import('@/views/openfinance/DashboardView')));
const AccountsView = Loadable(lazy(() => import('@/views/openfinance/AccountsView')));
const ConsentsView = Loadable(lazy(() => import('@/views/openfinance/ConsentsView')));
const TransactionsView = Loadable(lazy(() => import('@/views/openfinance/TransactionsView')));

const MainRoutes = {
  path: '/',
  element: (
    <AuthProvider>
      <AuthGuard>
        <RoleGuard>
          <AdminLayout />
        </RoleGuard>
      </AuthGuard>
    </AuthProvider>
  ),
  children: [
    { index: true, element: <Navigate to="/openfinance" replace /> },
    { path: 'openfinance', element: <OpenFinanceDashboard /> },
    { path: 'openfinance/accounts', element: <AccountsView /> },
    { path: 'openfinance/consents', element: <ConsentsView /> },
    { path: 'openfinance/transactions', element: <TransactionsView /> }
  ]
};

export default MainRoutes;
