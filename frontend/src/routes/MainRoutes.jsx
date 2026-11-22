import { lazy } from 'react';
import { Navigate } from 'react-router-dom';

// @project
import Loadable from '@/components/Loadable';
import AdminLayout from '@/layouts/AdminLayout';

// Views
const OpenFinanceDashboard = Loadable(lazy(() => import('@/views/openfinance/DashboardView')));
const AccountsView = Loadable(lazy(() => import('@/views/openfinance/AccountsView')));
const ConsentsView = Loadable(lazy(() => import('@/views/openfinance/ConsentsView')));
const TransactionsView = Loadable(lazy(() => import('@/views/openfinance/TransactionsView')));

const MainRoutes = {
  path: '/',
  element: <AdminLayout />,
  children: [
    { index: true, element: <Navigate to="/openfinance" replace /> },
    { path: 'openfinance', element: <OpenFinanceDashboard /> },
    { path: 'openfinance/accounts', element: <AccountsView /> },
    { path: 'openfinance/consents', element: <ConsentsView /> },
    { path: 'openfinance/transactions', element: <TransactionsView /> },
    { path: '*', element: <Navigate to="/openfinance" replace /> }
  ]
};

export default MainRoutes;
