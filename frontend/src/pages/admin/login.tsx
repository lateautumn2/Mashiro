import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { Activity, LogIn } from 'lucide-react';
import { apiRequest } from '../../lib/api';
import { isAuthenticated, saveAuth } from '../../lib/auth';

interface LoginResponse {
  token: string;
  user: string;
}

function getRedirectTarget(search: string) {
  const params = new URLSearchParams(search);
  const redirect = params.get('redirect');
  if (!redirect || !redirect.startsWith('/')) {
    return '/admin';
  }
  return redirect;
}

export default function AdminLogin() {
  const navigate = useNavigate();
  const location = useLocation();
  const [username, setUsername] = useState('admin');
  const [password, setPassword] = useState('admin');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (isAuthenticated()) {
      navigate(getRedirectTarget(location.search), { replace: true });
    }
  }, [location.search, navigate]);

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const response = await apiRequest<LoginResponse>('/login', {
        method: 'POST',
        body: JSON.stringify({
          username: username.trim(),
          password,
        }),
      });

      saveAuth(response.token, response.user);
      navigate(getRedirectTarget(location.search), { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-4">
      <div className="w-full max-w-md rounded-2xl border bg-card text-card-foreground shadow-lg">
        <div className="border-b px-6 py-5">
          <div className="flex items-center gap-3">
            <Activity className="h-6 w-6 text-primary" />
            <div>
              <h1 className="text-xl font-semibold">Admin Login</h1>
              <p className="text-sm text-muted-foreground">登录后进入后台管理。</p>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 px-6 py-6">
          <div className="space-y-2">
            <label className="text-sm font-medium">Username</label>
            <input
              type="text"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              className="w-full rounded-md border bg-transparent px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary"
              autoComplete="username"
            />
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Password</label>
            <input
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              className="w-full rounded-md border bg-transparent px-3 py-2 focus:outline-none focus:ring-2 focus:ring-primary"
              autoComplete="current-password"
            />
          </div>

          {error && (
            <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <button
            type="submit"
            disabled={isSubmitting}
            className="inline-flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90 disabled:opacity-50"
          >
            <LogIn className="mr-2 h-4 w-4" />
            {isSubmitting ? 'Signing in...' : 'Sign In'}
          </button>
        </form>
      </div>
    </div>
  );
}
