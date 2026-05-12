import { useState, useEffect } from 'react';
import { Save, Bell, Shield } from 'lucide-react';
import { apiRequest } from '../../lib/api';
import { getUsername } from '../../lib/auth';

export default function AdminSettings() {
  const [tgBotToken, setTgBotToken] = useState('');
  const [tgChatId, setTgChatId] = useState('');
  const [username] = useState(getUsername());
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchConfigs = async () => {
      setError('');
      try {
        const data = await apiRequest<Record<string, string>>('/admin/config', {}, { auth: true });
        if (data.tg_bot_token) setTgBotToken(data.tg_bot_token);
        if (data.tg_chat_id) setTgChatId(data.tg_chat_id);
      } catch (err) {
        const message = err instanceof Error ? err.message : '加载配置失败';
        setError(message);
      }
    };
    void fetchConfigs();
  }, []);

  const saveConfig = async (key: string, value: string) => {
    await apiRequest('/admin/config', {
      method: 'POST',
      body: JSON.stringify({ key, value })
    }, { auth: true });
  };

  const handleSaveTelegram = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await saveConfig('tg_bot_token', tgBotToken);
      await saveConfig('tg_chat_id', tgChatId);
      alert('Settings saved successfully');
    } catch (err) {
      const message = err instanceof Error ? err.message : '保存设置失败';
      setError(message);
    }
  };

  const handleSaveAccount = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await apiRequest('/admin/password', {
        method: 'POST',
        body: JSON.stringify({
          old_password: password,
          new_password: newPassword
        })
      }, { auth: true });
      alert('Password updated successfully');
      setPassword('');
      setNewPassword('');
    } catch (err) {
      const message = err instanceof Error ? err.message : '更新密码失败';
      setError(message);
    }
  };

  return (
    <div className="space-y-6 max-w-4xl">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground">Manage admin account and notifications.</p>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="grid gap-6">
        {/* Telegram Bot Settings */}
        <div className="rounded-xl border bg-card text-card-foreground shadow">
          <div className="p-6 border-b">
            <div className="flex items-center space-x-2">
              <Bell className="w-5 h-5 text-primary" />
              <h3 className="font-semibold text-lg">Telegram Notifications</h3>
            </div>
            <p className="text-sm text-muted-foreground mt-1">Configure Telegram bot for server offline alerts.</p>
          </div>
          <form onSubmit={handleSaveTelegram} className="p-6 space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Bot Token</label>
              <input
                type="text"
                value={tgBotToken}
                onChange={(e) => setTgBotToken(e.target.value)}
                placeholder="1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Chat ID</label>
              <input
                type="text"
                value={tgChatId}
                onChange={(e) => setTgChatId(e.target.value)}
                placeholder="-1001234567890"
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <button
              type="submit"
              className="inline-flex items-center justify-center rounded-md text-sm font-medium bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
            >
              <Save className="w-4 h-4 mr-2" /> Save Changes
            </button>
          </form>
        </div>

        {/* Account Settings */}
        <div className="rounded-xl border bg-card text-card-foreground shadow">
          <div className="p-6 border-b">
            <div className="flex items-center space-x-2">
              <Shield className="w-5 h-5 text-primary" />
              <h3 className="font-semibold text-lg">Admin Account</h3>
            </div>
            <p className="text-sm text-muted-foreground mt-1">Update your login credentials.</p>
          </div>
          <form onSubmit={handleSaveAccount} className="p-6 space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Username</label>
              <input
                type="text"
                value={username}
                readOnly
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Current Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">New Password</label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <button
              type="submit"
              className="inline-flex items-center justify-center rounded-md text-sm font-medium bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2"
            >
              <Save className="w-4 h-4 mr-2" /> Update Account
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
