import { useCallback, useEffect, useState } from 'react';
import { Bell, BellOff, AlertTriangle, Clock, Power } from 'lucide-react';
import { apiRequest } from '../../lib/api';

interface NotificationPref {
  server_id: number;
  server_name: string;
  notify_online: boolean;
  notify_offline: boolean;
  notify_traffic: boolean;
  notify_expiry: boolean;
}

type PrefKey = 'notify_online' | 'notify_offline' | 'notify_traffic' | 'notify_expiry';

export default function AdminNotifications() {
  const [prefs, setPrefs] = useState<NotificationPref[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [savingId, setSavingId] = useState<number | null>(null);
  const [configMissing, setConfigMissing] = useState(false);

  const loadPrefs = useCallback(async () => {
    setIsLoading(true);
    setError('');
    try {
      const data = await apiRequest<NotificationPref[]>(
        '/admin/notification-prefs',
        {},
        { auth: true }
      );
      setPrefs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load notification settings');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadPrefs();
    checkConfig();
  }, [loadPrefs]);

  const checkConfig = async () => {
    try {
      const config = await apiRequest<Record<string, string>>('/admin/config', {}, { auth: true });
      if (!config.tg_bot_token || !config.tg_chat_id) {
        setConfigMissing(true);
      }
    } catch {
      // If config check fails, assume it may be missing.
      setConfigMissing(true);
    }
  };

  const togglePref = async (serverId: number, key: PrefKey, current: boolean) => {
    setSavingId(serverId);
    try {
      const body: Record<string, boolean> = {};
      body[key] = !current;

      await apiRequest(
        `/admin/notification-prefs/${serverId}`,
        {
          method: 'PUT',
          body: JSON.stringify(body),
        },
        { auth: true }
      );

      setPrefs((prev) =>
        prev.map((p) =>
          p.server_id === serverId ? { ...p, [key]: !current } : p
        )
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update preference');
    } finally {
      setSavingId(null);
    }
  };

  const ToggleButton = ({
    serverId,
    prefKey,
    value,
  }: {
    serverId: number;
    prefKey: PrefKey;
    value: boolean;
  }) => (
    <button
      onClick={() => togglePref(serverId, prefKey, value)}
      disabled={savingId === serverId}
      className={`inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors ${
        value
          ? 'bg-primary/10 text-primary hover:bg-primary/20'
          : 'bg-muted text-muted-foreground hover:bg-muted/80'
      } disabled:opacity-50`}
    >
      {value ? (
        <Bell className="h-3.5 w-3.5" />
      ) : (
        <BellOff className="h-3.5 w-3.5" />
      )}
      {value ? 'On' : 'Off'}
    </button>
  );

  return (
    <div className="space-y-6 max-w-5xl">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Notifications</h1>
        <p className="text-muted-foreground">
          Configure per-server notification preferences for online/offline events, traffic warnings, and expiry alerts.
        </p>
      </div>

      {configMissing && (
        <div className="rounded-md border border-yellow-500/30 bg-yellow-500/10 px-4 py-3 text-sm">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4 text-yellow-500 flex-shrink-0" />
            <span>
              Telegram bot token or chat ID is not configured. Notifications will not be sent until you configure them in{' '}
              <a href="/admin/settings" className="underline font-medium">Settings</a>.
            </span>
          </div>
        </div>
      )}

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="p-8 text-center text-muted-foreground">Loading notification preferences...</div>
      ) : error ? (
        <div className="rounded-xl border bg-card text-card-foreground shadow p-8 text-center">
          <p className="text-muted-foreground">Failed to load notification preferences.</p>
          <button
            onClick={loadPrefs}
            className="mt-3 inline-flex items-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
          >
            Retry
          </button>
        </div>
      ) : prefs.length === 0 ? (
        <div className="rounded-xl border bg-card text-card-foreground shadow p-8 text-center text-muted-foreground">
          No servers found. Add servers in the Servers page first.
        </div>
      ) : (
        <div className="rounded-xl border bg-card text-card-foreground shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground">
                <tr>
                  <th className="px-6 py-4 font-medium">Server</th>
                  <th className="px-6 py-4 font-medium">
                    <div className="flex items-center gap-1.5">
                      <Power className="h-4 w-4" />
                      Online/Offline
                    </div>
                  </th>
                  <th className="px-6 py-4 font-medium">
                    <div className="flex items-center gap-1.5">
                      <AlertTriangle className="h-4 w-4" />
                      Traffic (85%)
                    </div>
                  </th>
                  <th className="px-6 py-4 font-medium">
                    <div className="flex items-center gap-1.5">
                      <Clock className="h-4 w-4" />
                      Expiry (15d)
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {prefs.map((pref) => (
                  <tr key={pref.server_id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-6 py-4 font-medium">{pref.server_name}</td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <ToggleButton
                          serverId={pref.server_id}
                          prefKey="notify_online"
                          value={pref.notify_online}
                        />
                        <span className="text-xs text-muted-foreground">Online</span>
                        <ToggleButton
                          serverId={pref.server_id}
                          prefKey="notify_offline"
                          value={pref.notify_offline}
                        />
                        <span className="text-xs text-muted-foreground">Offline</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <ToggleButton
                        serverId={pref.server_id}
                        prefKey="notify_traffic"
                        value={pref.notify_traffic}
                      />
                    </td>
                    <td className="px-6 py-4">
                      <ToggleButton
                        serverId={pref.server_id}
                        prefKey="notify_expiry"
                        value={pref.notify_expiry}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
        <h3 className="font-semibold text-lg mb-3">Notification Thresholds</h3>
        <div className="grid gap-4 text-sm text-muted-foreground md:grid-cols-3">
          <div className="flex items-start gap-2">
            <Power className="h-4 w-4 mt-0.5 text-green-500" />
            <div>
              <p className="font-medium text-foreground">Online/Offline</p>
              <p>Notifies when a server comes online or goes offline (20s timeout).</p>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <AlertTriangle className="h-4 w-4 mt-0.5 text-yellow-500" />
            <div>
              <p className="font-medium text-foreground">Traffic Warning</p>
              <p>Sends a one-time alert when monthly traffic exceeds 85% of the limit. Resets each billing cycle.</p>
            </div>
          </div>
          <div className="flex items-start gap-2">
            <Clock className="h-4 w-4 mt-0.5 text-blue-500" />
            <div>
              <p className="font-medium text-foreground">Expiry Warning</p>
              <p>Notifies daily when a server is within 15 days of its expiry date.</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
