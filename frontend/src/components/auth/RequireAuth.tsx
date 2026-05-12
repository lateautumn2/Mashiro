import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { isAuthenticated, isTokenExpired, clearAuth } from '../../lib/auth';

export function RequireAuth() {
  const location = useLocation();

  if (!isAuthenticated()) {
    const redirect = encodeURIComponent(`${location.pathname}${location.search}${location.hash}`);
    return <Navigate to={`/admin/login?redirect=${redirect}`} replace />;
  }

  if (isTokenExpired()) {
    clearAuth();
    return <Navigate to="/admin/login" replace />;
  }

  return <Outlet />;
}
