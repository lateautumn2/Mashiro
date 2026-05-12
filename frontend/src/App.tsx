import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { MainLayout } from './layouts/MainLayout';
import { AdminLayout } from './layouts/AdminLayout';
import { RequireAuth } from './components/auth/RequireAuth';
import Dashboard from './pages/dashboard';
import AdminServers from './pages/admin/servers';
import AdminLatency from './pages/admin/latency';
import AdminSettings from './pages/admin/settings';
import AdminNotifications from './pages/admin/notifications';
import AdminLogin from './pages/admin/login';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<MainLayout />}>
          <Route index element={<Dashboard />} />
        </Route>

        <Route path="/admin/login" element={<AdminLogin />} />

        <Route element={<RequireAuth />}>
          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<AdminServers />} />
            <Route path="latency" element={<AdminLatency />} />
            <Route path="settings" element={<AdminSettings />} />
            <Route path="notifications" element={<AdminNotifications />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
