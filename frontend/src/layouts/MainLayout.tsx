import { Outlet, Link, useLocation } from 'react-router-dom';
import { Activity, Settings } from 'lucide-react';
import { cn } from '../lib/utils';
import { isAuthenticated } from '../lib/auth';

export function MainLayout() {
  const location = useLocation();
  const authenticated = isAuthenticated();
  const adminLink = authenticated ? '/admin' : '/admin/login';

  return (
    <div className="min-h-screen bg-background flex flex-col">
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="container mx-auto flex h-14 items-center">
          <div className="mr-4 hidden md:flex">
            <Link to="/" className="mr-6 flex items-center space-x-2">
              <Activity className="h-6 w-6 text-primary" />
              <span className="hidden font-bold sm:inline-block">
                Mashiro Monitor
              </span>
            </Link>
            <nav className="flex items-center space-x-6 text-sm font-medium">
              <Link
                to="/"
                className={cn(
                  "transition-colors hover:text-foreground/80",
                  location.pathname === "/" ? "text-foreground" : "text-foreground/60"
                )}
              >
                Dashboard
              </Link>
            </nav>
          </div>
          <div className="flex flex-1 items-center justify-between space-x-2 md:justify-end">
            <div className="w-full flex-1 md:w-auto md:flex-none">
            </div>
            <nav className="flex items-center">
              <Link to={adminLink}>
                <div className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-9 py-2 px-3">
                  <Settings className={cn('h-4 w-4', authenticated ? 'mr-2' : '')} />
                  {authenticated ? 'Admin' : null}
                </div>
              </Link>
            </nav>
          </div>
        </div>
      </header>
      <main className="flex-1 container mx-auto py-6">
        <Outlet />
      </main>
    </div>
  );
}
