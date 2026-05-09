import type { ReactNode } from 'react';
import { Activity, Clock3, Cpu, HardDrive, MapPin, MemoryStick as Memory, Monitor, Network, Wifi } from 'lucide-react';
import { cn } from '../../lib/utils';
import type { Server } from '../../store/useStore';

type ServerItem = Server;

const TILE_CLASS = 'rounded-2xl bg-slate-50/75 px-3 py-2.5 dark:bg-slate-950/40';

function formatBytes(bytes: number) {
  if (bytes === 0) return '0 B';
  const base = 1024;
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  const index = Math.floor(Math.log(bytes) / Math.log(base));
  return `${parseFloat((bytes / Math.pow(base, index)).toFixed(1))} ${units[index]}`;
}

function formatDuration(seconds: number) {
  if (seconds <= 0) return '0m';
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);

  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}

function formatRemainingTime(expiryTime?: string) {
  if (!expiryTime) return 'Permanent';
  const expiryAt = new Date(expiryTime);
  if (Number.isNaN(expiryAt.getTime())) return '--';

  const remainingSeconds = Math.floor((expiryAt.getTime() - Date.now()) / 1000);
  if (remainingSeconds <= 0) return 'Expired';
  return formatDuration(remainingSeconds);
}

function formatRegionLabel(region?: string) {
  if (!region) return 'Unknown Region';
  const parts = region.split('·').map((part) => part.trim()).filter(Boolean);
  if (parts.length >= 2 && /^[A-Z]{2,3}$/.test(parts[0])) {
    return parts.slice(1).join(' · ');
  }
  return region;
}

function formatLatency(latencyMs: number) {
  if (latencyMs < 10) {
    return `${latencyMs.toFixed(2)} ms`;
  }
  if (latencyMs < 100) {
    return `${latencyMs.toFixed(1)} ms`;
  }
  return `${latencyMs.toFixed(0)} ms`;
}

function getUsagePercent(used: number, total: number, isOnline: boolean) {
  if (!isOnline || total === 0) return 0;
  return Math.round((used / total) * 100);
}

function getProgressTone(percent: number) {
  if (percent > 80) return 'bg-rose-500';
  if (percent > 60) return 'bg-yellow-500';
  return 'bg-emerald-500';
}

function getLatencyTone(value: number, status: string, isOnline: boolean) {
  if (!isOnline || status !== 'ok') return 'bg-slate-200 dark:bg-slate-800';
  if (value <= 60) return 'bg-emerald-500';
  if (value <= 120) return 'bg-yellow-500';
  if (value <= 180) return 'bg-orange-500';
  return 'bg-rose-500';
}

function getLatencyBars(value: number, status: string, isOnline: boolean) {
  if (!isOnline || status !== 'ok') return 0;
  if (value <= 60) return 5;
  if (value <= 120) return 4;
  if (value <= 180) return 3;
  if (value <= 260) return 2;
  return 1;
}

function getTrafficUsagePercent(totalUsed: number, limit: number) {
  if (limit === 0) return 0;
  return Math.min(100, Math.round((totalUsed / limit) * 100));
}

function InfoChip({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5 text-[10px] font-medium text-slate-700 dark:border-slate-800 dark:bg-slate-900 dark:text-slate-300">
      {icon}
      {label}
    </span>
  );
}

function UsageTile({
  icon,
  title,
  value,
  detail,
  percent,
}: {
  icon: ReactNode;
  title: string;
  value: string;
  detail: string;
  percent: number;
}) {
  return (
    <div className={cn(TILE_CLASS, 'space-y-1.5')}>
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 text-xs font-medium text-slate-700 dark:text-slate-300">
          {icon}
          <span>{title}</span>
        </div>
        <span className="text-base font-semibold leading-none">{value}</span>
      </div>
      <div className="text-xs text-muted-foreground">{detail}</div>
      <div className="h-1.5 overflow-hidden rounded-full bg-secondary">
        <div className={cn('h-full rounded-full transition-all', getProgressTone(percent))} style={{ width: `${percent}%` }} />
      </div>
    </div>
  );
}

function TrafficTile({ server, isOnline }: { server: ServerItem; isOnline: boolean }) {
  const totalTraffic = server.network.totalRx + server.network.totalTx;
  const trafficLimit = server.trafficLimit ?? 0;
  const progress = getTrafficUsagePercent(totalTraffic, trafficLimit);
  const value = trafficLimit > 0 ? formatBytes(trafficLimit) : 'Unlimited';
  const detail = trafficLimit > 0
    ? `已用 ${isOnline ? formatBytes(totalTraffic) : '0 B'}`
    : '未设置流量上限';

  return (
    <div className={cn(TILE_CLASS, 'space-y-1.5')}>
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1.5 text-[11px] font-medium text-slate-700 dark:text-slate-300">
          <Activity className="h-3.5 w-3.5 text-violet-500" />
          <span>Traffic</span>
        </div>
        <span className="text-sm font-semibold leading-none">{value}</span>
      </div>
      <div className="text-[11px] text-muted-foreground">{detail}</div>
      <div className="h-1.5 overflow-hidden rounded-full bg-secondary">
        <div className={cn('h-full rounded-full transition-all', trafficLimit > 0 ? getProgressTone(progress) : 'bg-slate-300 dark:bg-slate-700')} style={{ width: `${trafficLimit > 0 ? progress : 100}%` }} />
      </div>
    </div>
  );
}

function SpeedTile({ server, isOnline }: { server: ServerItem; isOnline: boolean }) {
  const upText = isOnline ? `${formatBytes(server.network.tx)}/s` : '-';
  const downText = isOnline ? `${formatBytes(server.network.rx)}/s` : '-';

  return (
    <div className={cn(TILE_CLASS, 'flex h-full flex-col justify-between space-y-2')}>
      <div className="flex items-center gap-1.5 text-xs font-medium text-slate-700 dark:text-slate-300">
        <Network className="h-3.5 w-3.5 text-violet-500" />
        <span>Speed</span>
      </div>
      <div className="space-y-1.5 text-sm">
        <div className="flex items-center justify-between rounded-xl bg-emerald-50/90 px-2.5 py-1.5 dark:bg-emerald-950/20">
          <span className="text-slate-600 dark:text-slate-300">Up</span>
          <span className="font-semibold text-emerald-600 dark:text-emerald-400">{upText}</span>
        </div>
        <div className="flex items-center justify-between rounded-xl bg-sky-50/90 px-2.5 py-1.5 dark:bg-sky-950/20">
          <span className="text-slate-600 dark:text-slate-300">Down</span>
          <span className="font-semibold text-sky-600 dark:text-sky-400">{downText}</span>
        </div>
      </div>
    </div>
  );
}

function LatencyCard({
  isOnline,
  latencyResults,
  uptime,
  expiryTime,
}: {
  isOnline: boolean;
  latencyResults: ServerItem['latencyResults'];
  uptime: number;
  expiryTime?: string;
}) {
  const hasLatencyResults = latencyResults.length > 0;

  return (
    <div className="rounded-2xl border border-border/60 bg-white/85 px-3.5 py-2.5 dark:bg-slate-900/70">
      <div className="space-y-2">
        <div className="inline-flex items-center gap-1.5 text-xs font-medium text-slate-700 dark:text-slate-300">
          <Wifi className="h-3.5 w-3.5" />
          <span>Latency Tasks</span>
        </div>
        {hasLatencyResults ? (
          <div className="space-y-2">
            {latencyResults.map((item) => {
              const latencyTone = getLatencyTone(item.latencyMs, item.status, isOnline);
              const latencyBars = getLatencyBars(item.latencyMs, item.status, isOnline);

              return (
                <div key={item.taskId} className="rounded-xl bg-slate-50 px-2.5 py-2 dark:bg-slate-950/40">
                  <div className="flex items-center justify-between gap-3 text-xs">
                    <div className="font-medium text-foreground">{item.taskName}</div>
                    <span className="text-sm font-semibold">
                      {item.status === 'ok' ? formatLatency(item.latencyMs) : 'Error'}
                    </span>
                  </div>
                  <div className="mt-2 flex items-end gap-1">
                    {[1, 2, 3, 4, 5].map((bar) => (
                      <span
                        key={bar}
                        className={cn('h-1.5 flex-1 rounded-full transition-all', bar <= latencyBars ? latencyTone : 'bg-slate-200 dark:bg-slate-800')}
                      />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="rounded-xl bg-slate-50 px-2.5 py-2 text-xs text-muted-foreground dark:bg-slate-950/40">
            暂无延迟任务结果
          </div>
        )}
      </div>
      <div className="mt-2 grid gap-2 sm:grid-cols-2">
        <div className="rounded-xl bg-slate-50 px-2.5 py-2 dark:bg-slate-950/40">
          <div className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
            <Clock3 className="h-3 w-3" />
            剩余时间
          </div>
          <div className="mt-1 text-sm font-semibold">{formatRemainingTime(expiryTime)}</div>
        </div>
        <div className="rounded-xl bg-slate-50 px-2.5 py-2 dark:bg-slate-950/40">
          <div className="inline-flex items-center gap-1 text-[11px] text-muted-foreground">
            <Activity className="h-3 w-3" />
            运行时间
          </div>
          <div className="mt-1 text-sm font-semibold">{isOnline ? formatDuration(uptime) : '-'}</div>
        </div>
      </div>
    </div>
  );
}

function ServerCardHeader({ server, isOnline }: { server: ServerItem; isOnline: boolean }) {
  const systemLabel = server.system || 'Unknown';
  const regionLabel = formatRegionLabel(server.region);

  return (
    <div className="border-b border-border/50 px-3.5 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex items-center gap-2">
          <div className={cn('h-2.5 w-2.5 rounded-full', isOnline ? 'bg-emerald-500 shadow-[0_0_10px_rgba(34,197,94,0.6)]' : 'bg-rose-500 shadow-[0_0_10px_rgba(239,68,68,0.6)]')} />
          <div>
            <h3 className="text-base font-semibold leading-tight">{server.name}</h3>
            <div className="mt-0.5 text-xs text-muted-foreground">{isOnline ? '在线运行中' : '当前离线'}</div>
          </div>
        </div>
        <div className={cn('rounded-full px-2 py-0.5 text-[10px] font-medium', isOnline ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300' : 'bg-rose-100 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300')}>
          {isOnline ? 'ONLINE' : 'OFFLINE'}
        </div>
      </div>
      <div className="mt-2 flex flex-wrap gap-1.5">
        <InfoChip icon={<MapPin className="h-3 w-3" />} label={regionLabel} />
        <InfoChip icon={<Monitor className="h-3 w-3" />} label={systemLabel} />
        <InfoChip icon={<span className={cn('h-1.5 w-1.5 rounded-full', server.hasIPv4 ? 'bg-sky-500' : 'bg-slate-300 dark:bg-slate-600')} />} label="IPv4" />
        <InfoChip icon={<span className={cn('h-1.5 w-1.5 rounded-full', server.hasIPv6 ? 'bg-violet-500' : 'bg-slate-300 dark:bg-slate-600')} />} label="IPv6" />
      </div>
    </div>
  );
}

export function ServerCard({ server }: { server: ServerItem }) {
  const isOnline = server.status === 'online';
  const memoryPercent = getUsagePercent(server.memory.used, server.memory.total, isOnline);
  const diskPercent = getUsagePercent(server.disk.used, server.disk.total, isOnline);
  const cpuPercent = isOnline ? Math.round(server.cpu) : 0;

  return (
    <div className="group rounded-3xl border border-white/60 bg-white/90 text-card-foreground shadow-[0_14px_48px_rgba(15,23,42,0.08)] backdrop-blur-xl transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_20px_72px_rgba(59,130,246,0.14)] dark:border-white/10 dark:bg-slate-900/80">
      <ServerCardHeader server={server} isOnline={isOnline} />
      <div className="space-y-2.5 px-3.5 py-3">
        <div className="grid gap-2 sm:grid-cols-2 sm:grid-rows-[auto_auto_auto]">
          <UsageTile
            icon={<Cpu className="h-3.5 w-3.5 text-violet-500" />}
            title="CPU"
            value={isOnline ? `${server.cpu.toFixed(1)}%` : '-'}
            detail={isOnline ? `${server.cpuCores} Cores` : '-'}
            percent={cpuPercent}
          />
          <UsageTile
            icon={<Memory className="h-3.5 w-3.5 text-violet-500" />}
            title="RAM"
            value={isOnline ? `${memoryPercent}%` : '-'}
            detail={isOnline ? `${formatBytes(server.memory.used)} / ${formatBytes(server.memory.total)}` : '-'}
            percent={memoryPercent}
          />
          <UsageTile
            icon={<HardDrive className="h-3.5 w-3.5 text-violet-500" />}
            title="Disk"
            value={isOnline ? `${diskPercent}%` : '-'}
            detail={isOnline ? `${formatBytes(server.disk.used)} / ${formatBytes(server.disk.total)}` : '-'}
            percent={diskPercent}
          />
          <div className="sm:row-span-2">
            <SpeedTile server={server} isOnline={isOnline} />
          </div>
          <TrafficTile server={server} isOnline={isOnline} />
        </div>
        <LatencyCard isOnline={isOnline} latencyResults={server.latencyResults} uptime={server.uptime} expiryTime={server.expiryTime} />
      </div>
    </div>
  );
}
