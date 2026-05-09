import { useCallback, useEffect, useState } from 'react';
import { AddServerModal, type AddServerFormValues } from '../../components/ui/AddServerModal';
import { DeployCommandModal, type DeployCommands } from '../../components/ui/DeployCommandModal';
import { Pencil, Plus, Trash2, TerminalSquare } from 'lucide-react';
import { apiRequest } from '../../lib/api';
import { useStore } from '../../store/useStore';

interface AdminServerResponse {
  id: number;
  name: string;
  ip: string;
  region: string;
  system: string;
  has_ipv4: boolean;
  has_ipv6: boolean;
  status: 'online' | 'offline';
  traffic_limit: number;
  traffic_reset_day: number;
  expiry_time: string;
}

interface AdminServer {
  id: number;
  name: string;
  ip: string;
  region: string;
  system: string;
  hasIPv4: boolean;
  hasIPv6: boolean;
  status: 'online' | 'offline';
  trafficLimit: number;
  trafficResetDay: number;
  expiryTime: string;
}

function mapAdminServer(server: AdminServerResponse): AdminServer {
  return {
    id: server.id,
    name: server.name,
    ip: server.ip,
    region: server.region,
    system: server.system,
    hasIPv4: server.has_ipv4,
    hasIPv6: server.has_ipv6,
    status: server.status,
    trafficLimit: server.traffic_limit,
    trafficResetDay: server.traffic_reset_day,
    expiryTime: server.expiry_time,
  };
}

function formatTrafficLimit(bytes: number) {
  if (!bytes) {
    return 'Unlimited';
  }

  return `${(bytes / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

function formatExpiryTime(value: string) {
  if (!value || value.startsWith('0001-01-01')) {
    return '-';
  }

  return new Date(value).toLocaleDateString();
}

export default function AdminServers() {
  const { fetchServers } = useStore();
  const [servers, setServers] = useState<AdminServer[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<AdminServer | null>(null);
  const [deployCommands, setDeployCommands] = useState<DeployCommands | null>(null);
  const [deployTitle, setDeployTitle] = useState('一键部署指令');
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [commandLoadingId, setCommandLoadingId] = useState<number | null>(null);

  const loadServers = useCallback(async () => {
    setError('');
    setIsLoading(true);

    try {
      const data = await apiRequest<AdminServerResponse[]>('/admin/servers', {}, { auth: true });
      setServers(data.map(mapAdminServer));
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载服务器列表失败');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadServers();
  }, [loadServers]);

  const refreshDashboardServers = async () => {
    try {
      await fetchServers();
    } catch (err) {
      console.error(err);
    }
  };

  const handleAddServer = async (values: AddServerFormValues) => {
    const createdServer = await apiRequest<AdminServerResponse>(
      '/admin/servers',
      {
        method: 'POST',
        body: JSON.stringify({
          name: values.name,
          traffic_limit: values.trafficLimit,
          traffic_reset_day: values.trafficResetDay,
          expiry_time: values.expiryTime,
        }),
      },
      { auth: true },
    );
    const commandResponse = await apiRequest<DeployCommands>(
      `/admin/servers/${createdServer.id}/command`,
      {},
      { auth: true },
    );

    await loadServers();
    await refreshDashboardServers();
    return commandResponse;
  };

  const handleEditServer = async (values: AddServerFormValues) => {
    if (!editingServer) {
      return;
    }

    await apiRequest<AdminServerResponse>(
      `/admin/servers/${editingServer.id}`,
      {
        method: 'PUT',
        body: JSON.stringify({
          name: values.name,
          traffic_limit: values.trafficLimit,
          traffic_reset_day: values.trafficResetDay,
          expiry_time: values.expiryTime,
        }),
      },
      { auth: true },
    );

    await loadServers();
    await refreshDashboardServers();
    setEditingServer(null);
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确认删除该服务器吗？')) {
      return;
    }

    setError('');

    try {
      await apiRequest(`/admin/servers/${id}`, { method: 'DELETE' }, { auth: true });
      await loadServers();
      await refreshDashboardServers();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除服务器失败');
    }
  };

  const handleShowCommand = async (id: number) => {
    setCommandLoadingId(id);
    setError('');

    try {
      const commands = await apiRequest<DeployCommands>(
        `/admin/servers/${id}/command`,
        {},
        { auth: true },
      );
      const server = servers.find((item) => item.id === id);
      setDeployTitle(server ? `${server.name} 的一键部署指令` : '一键部署指令');
      setDeployCommands(commands);
    } catch (err) {
      setError(err instanceof Error ? err.message : '获取部署命令失败');
    } finally {
      setCommandLoadingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Servers</h1>
          <p className="text-muted-foreground">Manage your monitored servers.</p>
        </div>
        <button
          onClick={() => {
            setEditingServer(null);
            setIsModalOpen(true);
          }}
          className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground shadow hover:bg-primary/90 h-9 px-4 py-2 w-full sm:w-auto"
        >
          <Plus className="mr-2 h-4 w-4" /> Add Server
        </button>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="rounded-xl border bg-card text-card-foreground shadow overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm text-left">
            <thead className="bg-muted/50 text-muted-foreground">
              <tr>
                <th className="px-6 py-4 font-medium">ID</th>
                <th className="px-6 py-4 font-medium">Name</th>
                <th className="px-6 py-4 font-medium">Region</th>
                <th className="px-6 py-4 font-medium">System</th>
                <th className="px-6 py-4 font-medium">IP</th>
                <th className="px-6 py-4 font-medium">Status</th>
                <th className="px-6 py-4 font-medium">Traffic Limit</th>
                <th className="px-6 py-4 font-medium">Reset Day</th>
                <th className="px-6 py-4 font-medium">Expiry</th>
                <th className="px-6 py-4 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {servers.map((server) => (
                <tr key={server.id} className="hover:bg-muted/50 transition-colors">
                  <td className="px-6 py-4 font-mono text-xs text-muted-foreground">{server.id}</td>
                  <td className="px-6 py-4 font-medium">{server.name}</td>
                  <td className="px-6 py-4">{server.region || '-'}</td>
                  <td className="px-6 py-4">{server.system || '-'}</td>
                  <td className="px-6 py-4 font-mono">{server.ip || '-'}</td>
                  <td className="px-6 py-4">
                    <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${server.status === 'online' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'}`}>
                      {server.status === 'online' ? 'Online' : 'Offline'}
                    </span>
                  </td>
                  <td className="px-6 py-4">{formatTrafficLimit(server.trafficLimit)}</td>
                  <td className="px-6 py-4">Day {server.trafficResetDay}</td>
                  <td className="px-6 py-4">{formatExpiryTime(server.expiryTime)}</td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex items-center justify-end space-x-2">
                      <button
                        title="Edit Server"
                        className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
                        onClick={() => {
                          setEditingServer(server);
                          setIsModalOpen(true);
                        }}
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                      <button
                        title="View Install Command"
                        className="p-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-md transition-colors"
                        disabled={commandLoadingId === server.id}
                        onClick={() => void handleShowCommand(server.id)}
                      >
                        <TerminalSquare className="w-4 h-4" />
                      </button>
                      <button
                        title="Delete Server"
                        onClick={() => void handleDelete(server.id)}
                        className="p-2 text-destructive hover:bg-destructive/10 rounded-md transition-colors"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!isLoading && servers.length === 0 && (
            <div className="p-8 text-center text-muted-foreground">
              No servers found. Add one to get started.
            </div>
          )}
          {isLoading && (
            <div className="p-8 text-center text-muted-foreground">
              Loading servers...
            </div>
          )}
        </div>
      </div>

      <AddServerModal
        isOpen={isModalOpen}
        onClose={() => {
          setIsModalOpen(false);
          setEditingServer(null);
        }}
        mode={editingServer ? 'edit' : 'create'}
        initialValues={editingServer ? {
          name: editingServer.name,
          trafficLimit: editingServer.trafficLimit,
          trafficResetDay: editingServer.trafficResetDay,
          expiryTime: editingServer.expiryTime,
        } : null}
        onSubmit={editingServer ? handleEditServer : handleAddServer}
      />
      <DeployCommandModal
        isOpen={deployCommands !== null}
        commands={deployCommands}
        title={deployTitle}
        onClose={() => setDeployCommands(null)}
      />
    </div>
  );
}
