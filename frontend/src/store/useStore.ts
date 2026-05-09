import { create } from 'zustand';
import { apiRequest } from '../lib/api';

export interface Server {
  id: string;
  name: string;
  system?: string;
  region?: string;
  hasIPv4: boolean;
  hasIPv6: boolean;
  status: 'online' | 'offline';
  uptime: number;
  cpuCores: number;
  cpu: number;
  memory: {
    total: number;
    used: number;
  };
  disk: {
    total: number;
    used: number;
  };
  network: {
    rx: number;
    tx: number;
    totalRx: number;
    totalTx: number;
  };
  latencyResults: Array<{
    taskId: number;
    taskName: string;
    type: string;
    target: string;
    latencyMs: number;
    status: string;
    errorMessage: string;
    checkedAt: string;
  }>;
  latency?: number;
  trafficLimit?: number;
  expiryTime?: string;
}

interface DashboardServerResponse {
  id: number;
  name: string;
  status: 'online' | 'offline';
  region: string;
  system: string;
  has_ipv4: boolean;
  has_ipv6: boolean;
  traffic_limit: number;
  expiry_time: string;
  uptime: number;
  cpu_cores: number;
  cpu_usage: number;
  mem_total: number;
  mem_used: number;
  disk_total: number;
  disk_used: number;
  net_in: number;
  net_out: number;
  net_in_speed: number;
  net_out_speed: number;
  latency: number;
  latency_results: Array<{
    task_id: number;
    task_name: string;
    type: string;
    target: string;
    latency_ms: number;
    status: string;
    error_message: string;
    checked_at: string;
  }>;
}

interface AppState {
  servers: Server[];
  setServers: (servers: Server[]) => void;
  fetchServers: () => Promise<void>;
}

function mapDashboardServer(server: DashboardServerResponse): Server {
  return {
    id: String(server.id),
    name: server.name,
    region: server.region,
    system: server.system,
    hasIPv4: server.has_ipv4,
    hasIPv6: server.has_ipv6,
    trafficLimit: server.traffic_limit,
    expiryTime: server.expiry_time,
    status: server.status,
    uptime: server.uptime,
    cpuCores: server.cpu_cores,
    cpu: server.cpu_usage,
    memory: {
      total: server.mem_total,
      used: server.mem_used,
    },
    disk: {
      total: server.disk_total,
      used: server.disk_used,
    },
    network: {
      totalRx: server.net_in,
      totalTx: server.net_out,
      rx: server.net_in_speed,
      tx: server.net_out_speed,
    },
    latencyResults: server.latency_results.map((item) => ({
      taskId: item.task_id,
      taskName: item.task_name,
      type: item.type,
      target: item.target,
      latencyMs: item.latency_ms,
      status: item.status,
      errorMessage: item.error_message,
      checkedAt: item.checked_at,
    })),
    latency: server.latency,
  };
}

export const useStore = create<AppState>((set) => ({
  servers: [],
  setServers: (servers) => set({ servers }),
  fetchServers: async () => {
    const data = await apiRequest<DashboardServerResponse[]>('/dashboard/servers');
    set({ servers: data.map(mapDashboardServer) });
  },
}));
