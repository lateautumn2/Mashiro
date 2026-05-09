import { useEffect, useState } from 'react';
import { useStore } from '../../store/useStore';
import { ServerCard } from '../../components/ui/ServerCard';
import { ArrowDownToLine, ArrowUpToLine, LayoutGrid, List, RefreshCw, Router, Server as ServerIcon, ShieldCheck, ShieldX } from 'lucide-react';

const REFRESH_INTERVAL_MS = 5000;

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B';
  const base = 1024;
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.floor(Math.log(bytes) / Math.log(base));
  return `${parseFloat((bytes / Math.pow(base, index)).toFixed(1))} ${units[index]}`;
}

function formatRegionLabel(region?: string) {
  if (!region) return '-';
  const parts = region.split('·').map((part) => part.trim()).filter(Boolean);
  if (parts.length >= 2 && /^[A-Z]{2,3}$/.test(parts[0])) {
    return parts.slice(1).join(' · ');
  }
  return region;
}

function getTrafficUsagePercent(totalUsed: number, limit: number) {
  if (limit === 0) {
    return 0;
  }

  return Math.min(100, Math.round((totalUsed / limit) * 100));
}

function getProgressTone(percent: number) {
  if (percent > 80) return 'bg-rose-500';
  if (percent > 60) return 'bg-yellow-500';
  return 'bg-emerald-500';
}

function MetricCell({
  value,
  detail,
  percent,
}: {
  value: string;
  detail: string;
  percent: number;
}) {
  return (
    <div className="min-w-[150px] space-y-1.5">
      <div className="text-sm font-semibold">{value}</div>
      <div className="text-xs text-muted-foreground">{detail}</div>
      <div className="h-1.5 overflow-hidden rounded-full bg-secondary">
        <div className={`h-full rounded-full transition-all ${getProgressTone(percent)}`} style={{ width: `${percent}%` }} />
      </div>
    </div>
  );
}

function SummaryCard({
  title,
  value,
  detail,
  icon,
  tone,
}: {
  title: string;
  value: string;
  detail?: string;
  icon: React.ReactNode;
  tone: string;
}) {
  return (
    <div className="rounded-2xl border border-white/60 bg-white/90 p-5 text-card-foreground shadow-[0_12px_40px_rgba(15,23,42,0.07)] backdrop-blur-xl dark:border-white/10 dark:bg-slate-900/80">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-2">
          <div className="text-sm font-medium text-muted-foreground">{title}</div>
          <div className="text-3xl font-semibold tracking-tight">{value}</div>
          {detail ? <div className="text-xs text-muted-foreground">{detail}</div> : null}
        </div>
        <div className={`flex h-11 w-11 items-center justify-center rounded-2xl ${tone}`}>
          {icon}
        </div>
      </div>
    </div>
  );
}

export default function Dashboard() {
  const { servers, fetchServers } = useStore();
  const [view, setView] = useState<'grid' | 'table'>('grid');
  const [error, setError] = useState('');

  const getUsagePercent = (used: number, total: number) => {
    if (total === 0) {
      return 0;
    }

    return Math.round((used / total) * 100);
  };

  useEffect(() => {
    let active = true;

    const load = () => {
      setError('');
      fetchServers().catch((err: unknown) => {
        if (!active) {
          return;
        }
        setError(err instanceof Error ? err.message : '加载服务器列表失败');
      });
    };

    load();
    const timer = window.setInterval(load, REFRESH_INTERVAL_MS);

    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [fetchServers]);

  const onlineCount = servers.filter(s => s.status === 'online').length;
  const offlineCount = servers.length - onlineCount;
  const totalTrafficUsed = servers.reduce((sum, server) => sum + server.network.totalRx + server.network.totalTx, 0);
  const totalUpSpeed = servers.reduce((sum, server) => sum + (server.status === 'online' ? server.network.tx : 0), 0);
  const totalDownSpeed = servers.reduce((sum, server) => sum + (server.status === 'online' ? server.network.rx : 0), 0);

  return (
    <div className="space-y-8">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-end">
        <div className="inline-flex items-center gap-2 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-medium text-emerald-700 dark:border-emerald-900/40 dark:bg-emerald-950/20 dark:text-emerald-300">
          <RefreshCw className="h-3.5 w-3.5" />
          自动刷新中
        </div>
        <div className="flex items-center space-x-2 rounded-lg bg-muted p-1">
          <button
            onClick={() => setView('grid')}
            className={`p-2 rounded-md transition-colors ${view === 'grid' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}`}
          >
            <LayoutGrid className="w-4 h-4" />
          </button>
          <button
            onClick={() => setView('table')}
            className={`p-2 rounded-md transition-colors ${view === 'table' ? 'bg-background shadow-sm' : 'hover:bg-background/50'}`}
          >
            <List className="w-4 h-4" />
          </button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <SummaryCard
          title="Total Servers"
          value={String(servers.length)}
          detail="当前纳管服务器总数"
          icon={<ServerIcon className="h-5 w-5 text-indigo-600 dark:text-indigo-300" />}
          tone="bg-indigo-100 text-indigo-700 dark:bg-indigo-950/40"
        />
        <SummaryCard
          title="Online"
          value={String(onlineCount)}
          detail="在线服务器数量"
          icon={<ShieldCheck className="h-5 w-5 text-emerald-600 dark:text-emerald-300" />}
          tone="bg-emerald-100 text-emerald-700 dark:bg-emerald-950/40"
        />
        <SummaryCard
          title="Offline"
          value={String(offlineCount)}
          detail="离线服务器数量"
          icon={<ShieldX className="h-5 w-5 text-rose-600 dark:text-rose-300" />}
          tone="bg-rose-100 text-rose-700 dark:bg-rose-950/40"
        />
        <SummaryCard
          title="Traffic Used"
          value={formatBytes(totalTrafficUsed)}
          detail="所有服务器当期累计流量"
          icon={<Router className="h-5 w-5 text-violet-600 dark:text-violet-300" />}
          tone="bg-violet-100 text-violet-700 dark:bg-violet-950/40"
        />
        <SummaryCard
          title="Speed Summary"
          value={`${formatBytes(totalUpSpeed)}/s`}
          detail={`Down ${formatBytes(totalDownSpeed)}/s`}
          icon={
            <div className="flex items-center gap-1">
              <ArrowUpToLine className="h-4 w-4 text-emerald-600 dark:text-emerald-300" />
              <ArrowDownToLine className="h-4 w-4 text-sky-600 dark:text-sky-300" />
            </div>
          }
          tone="bg-sky-100 text-sky-700 dark:bg-sky-950/40"
        />
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {view === 'grid' ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {servers.map((server) => (
            <ServerCard key={server.id} server={server} />
          ))}
        </div>
      ) : (
        <div className="rounded-xl border bg-card text-card-foreground shadow overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm text-left">
              <thead className="bg-muted/50 text-muted-foreground">
                <tr>
                  <th className="px-6 py-4 font-medium">Server</th>
                  <th className="px-6 py-4 font-medium">Region</th>
                  <th className="px-6 py-4 font-medium">System</th>
                  <th className="px-6 py-4 font-medium">Status</th>
                  <th className="px-6 py-4 font-medium">CPU</th>
                  <th className="px-6 py-4 font-medium">RAM</th>
                  <th className="px-6 py-4 font-medium">Disk</th>
                  <th className="px-6 py-4 font-medium">Traffic</th>
                  <th className="px-6 py-4 font-medium">Speed</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {servers.map((server) => {
                  const isOnline = server.status === 'online';
                  const cpuPercent = isOnline ? Math.round(server.cpu) : 0;
                  const memoryPercent = isOnline ? getUsagePercent(server.memory.used, server.memory.total) : 0;
                  const diskPercent = isOnline ? getUsagePercent(server.disk.used, server.disk.total) : 0;
                  const totalTraffic = server.network.totalRx + server.network.totalTx;
                  const trafficLimit = server.trafficLimit ?? 0;
                  const trafficPercent = isOnline ? getTrafficUsagePercent(totalTraffic, trafficLimit) : 0;

                  return (
                    <tr key={server.id} className="hover:bg-muted/50 transition-colors">
                      <td className="px-6 py-4 font-medium">{server.name}</td>
                      <td className="px-6 py-4">{formatRegionLabel(server.region)}</td>
                      <td className="px-6 py-4">{server.system || '-'}</td>
                      <td className="px-6 py-4">
                        <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${isOnline ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'}`}>
                          {isOnline ? 'Online' : 'Offline'}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <MetricCell
                          value={isOnline ? `${server.cpu.toFixed(1)}%` : '-'}
                          detail={isOnline ? `${server.cpuCores} Cores` : '-'}
                          percent={cpuPercent}
                        />
                      </td>
                      <td className="px-6 py-4">
                        <MetricCell
                          value={isOnline ? `${memoryPercent}%` : '-'}
                          detail={isOnline ? `${formatBytes(server.memory.used)} / ${formatBytes(server.memory.total)}` : '-'}
                          percent={memoryPercent}
                        />
                      </td>
                      <td className="px-6 py-4">
                        <MetricCell
                          value={isOnline ? `${diskPercent}%` : '-'}
                          detail={isOnline ? `${formatBytes(server.disk.used)} / ${formatBytes(server.disk.total)}` : '-'}
                          percent={diskPercent}
                        />
                      </td>
                      <td className="px-6 py-4">
                        <MetricCell
                          value={trafficLimit > 0 ? formatBytes(trafficLimit) : 'Unlimited'}
                          detail={trafficLimit > 0 ? `已用 ${isOnline ? formatBytes(totalTraffic) : '0 B'}` : '未设置流量上限'}
                          percent={trafficLimit > 0 ? trafficPercent : 100}
                        />
                      </td>
                      <td className="px-6 py-4">
                        {isOnline ? (
                          <div className="min-w-[150px] space-y-1.5 text-xs">
                            <div className="flex items-center justify-between rounded-xl bg-emerald-50/90 px-2.5 py-2 dark:bg-emerald-950/20">
                              <span className="text-slate-600 dark:text-slate-300">Up</span>
                              <span className="font-semibold text-emerald-600 dark:text-emerald-400">{formatBytes(server.network.tx)}/s</span>
                            </div>
                            <div className="flex items-center justify-between rounded-xl bg-sky-50/90 px-2.5 py-2 dark:bg-sky-950/20">
                              <span className="text-slate-600 dark:text-slate-300">Down</span>
                              <span className="font-semibold text-sky-600 dark:text-sky-400">{formatBytes(server.network.rx)}/s</span>
                            </div>
                          </div>
                        ) : '-'}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
