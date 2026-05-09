import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Server, Settings, LogOut, Activity, Wifi } from 'lucide-react';
import { cn } from '../lib/utils';
import { clearAuth, getUsername } from '../lib/auth';

export function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();

  const navigation = [
    { name: 'Servers', href: '/admin', icon: Server },
    { name: 'Latency', href: '/admin/latency', icon: Wifi },
    { name: 'Settings', href: '/admin/settings', icon: Settings },
  ];

  const handleLogout = () => {
    clearAuth();
    navigate('/admin/login', { replace: true });
  };

  return (
    <div className="flex min-h-screen flex-col md:flex-row">
      {/* Sidebar */}
      <aside className="w-full md:w-64 border-r bg-muted/40 flex-shrink-0 flex flex-col">
        <div className="h-14 flex items-center border-b px-4">
          <Link to="/" className="flex items-center space-x-2 text-primary font-bold">
            <Activity className="h-5 w-5" />
            <span>Mashiro Admin</span>
          </Link>
        </div>
        <div className="border-b px-4 py-3 text-sm text-muted-foreground">
          当前登录：{getUsername()}
        </div>
        <div className="flex-1 py-4">
          <nav className="grid gap-1 px-2">
            {navigation.map((item) => {
              const isActive = location.pathname === item.href;
              return (
                <Link
                  key={item.name}
                  to={item.href}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-all",
                    isActive 
                      ? "bg-primary text-primary-foreground" 
                      : "text-muted-foreground hover:bg-muted hover:text-foreground"
                  )}
                >
                  <item.icon className="h-4 w-4" />
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="p-4 mt-auto">
          <button
            type="button"
            onClick={handleLogout}
            className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-muted-foreground hover:bg-muted transition-all"
          >
            <LogOut className="h-4 w-4" />
            Logout
          </button>
          <Link
            to="/"
            className="mt-2 flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-muted-foreground hover:bg-muted transition-all"
          >
            <LayoutDashboard className="h-4 w-4" />
            Back to Dashboard
          </Link>
        </div>
      </aside>
      
      {/* Main Content */}
      <main className="flex-1 p-6 md:p-8 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
