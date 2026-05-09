import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { isAuthenticated } from '../../lib/auth';

export function RequireAuth() {
  const location = useLocation();

  if (!isAuthenticated()) {
    const redirect = encodeURIComponent(`${location.pathname}${location.search}${location.hash}`);
    return <Navigate to={`/admin/login?redirect=${redirect}`} replace />;
  }

  return <Outlet />;
}
