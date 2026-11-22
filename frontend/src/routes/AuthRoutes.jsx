import { Navigate } from 'react-router-dom';

const AuthRoutes = {
  path: 'login',
  element: <Navigate to="/openfinance" replace />
};

export default AuthRoutes;
