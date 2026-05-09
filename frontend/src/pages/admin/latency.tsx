import { useCallback, useEffect, useMemo, useState } from 'react';
import { Pencil, Plus, Trash2 } from 'lucide-react';
import { apiRequest } from '../../lib/api';

interface AdminServerOption {
  id: number;
  name: string;
}

interface LatencyResult {
  task_id: number;
  server_id: number;
  task_name: string;
  type: string;
  target: string;
  latency_ms: number;
  status: string;
  error_message: string;
  checked_at: string;
}

interface LatencyTask {
  id: number;
  name: string;
  type: 'tcp' | 'icmp' | 'http';
  target: string;
  interval_sec: number;
  server_ids: number[];
  servers: AdminServerOption[];
  results: LatencyResult[];
}

interface TaskFormState {
  name: string;
  type: 'tcp' | 'icmp' | 'http';
  target: string;
  intervalSec: number;
  serverIDs: number[];
}

const defaultFormState: TaskFormState = {
  name: '',
  type: 'tcp',
  target: '',
  intervalSec: 60,
  serverIDs: [],
};

const HOST_LABEL_PATTERN = /^[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/;

function formatCheckTime(value: string) {
  if (!value) return '-';
  return new Date(value).toLocaleString();
}

function formatLatency(latencyMs: number) {
  if (latencyMs < 10) return `${latencyMs.toFixed(2)} ms`;
  if (latencyMs < 100) return `${latencyMs.toFixed(1)} ms`;
  return `${latencyMs.toFixed(0)} ms`;
}

function isValidHost(value: string) {
  const trimmed = value.trim().replace(/^\[|\]$/g, '');
  if (!trimmed || /[\/\\\s]/.test(trimmed)) return false;
  if (trimmed.toLowerCase() === 'localhost') return true;
  if (/^[0-9a-f:.]+$/i.test(trimmed) && trimmed.includes(':')) return true;
  if (/^\d{1,3}(?:\.\d{1,3}){3}$/.test(trimmed)) {
    return trimmed.split('.').every((segment) => {
      const value = Number(segment);
      return value >= 0 && value <= 255;
    });
  }
  return trimmed.split('.').every((label) => HOST_LABEL_PATTERN.test(label));
}

function validateTaskForm(form: TaskFormState) {
  const name = form.name.trim();
  const target = form.target.trim();

  if (!name) return '请填写任务名称';
  if (!target) return '请填写监测目标';
  if (form.serverIDs.length === 0) return '请至少选择一台执行监测的服务器';
  if (!Number.isFinite(form.intervalSec) || form.intervalSec < 5) return '监测间隔不能小于 5 秒';

  if (form.type === 'tcp') {
    const parts = target.match(/^\[([^\]]+)\]:(\d+)$/) ?? target.match(/^([^:]+):(\d+)$/);
    if (!parts) return 'TCP 目标必须填写为 host:port，例如 example.com:443';
    if (!isValidHost(parts[1])) return 'TCP 目标主机无效';
    const port = Number(parts[2]);
    if (!Number.isInteger(port) || port < 1 || port > 65535) return 'TCP 端口必须是 1-65535 之间的数字';
  }

  if (form.type === 'icmp' && !isValidHost(target)) {
    return 'ICMP 目标必须是合法的 IP 或主机名';
  }

  if (form.type === 'http') {
    try {
      const parsed = new URL(target);
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        return 'HTTP 目标必须以 http:// 或 https:// 开头';
      }
    } catch {
      return 'HTTP 目标必须是完整 URL，例如 https://example.com/health';
    }
  }

  return '';
}

function getTargetHint(type: TaskFormState['type']) {
  if (type === 'tcp') return 'TCP 必须填写 host:port，例如 1.2.3.4:443 或 example.com:443';
  if (type === 'http') return 'HTTP 必须填写完整 URL，例如 https://example.com/health';
  return 'ICMP 支持 IP 或主机名，例如 8.8.8.8 或 example.com';
}

export default function AdminLatency() {
  const [tasks, setTasks] = useState<LatencyTask[]>([]);
  const [servers, setServers] = useState<AdminServerOption[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isOpen, setIsOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<LatencyTask | null>(null);
  const [form, setForm] = useState<TaskFormState>(defaultFormState);
  const [error, setError] = useState('');

  const loadData = useCallback(async () => {
    setIsLoading(true);
    setError('');
    try {
      const [taskData, serverData] = await Promise.all([
        apiRequest<LatencyTask[]>('/admin/latency-tasks', {}, { auth: true }),
        apiRequest<Array<{ id: number; name: string }>>('/admin/servers', {}, { auth: true }),
      ]);
      setTasks(taskData);
      setServers(serverData.map((item) => ({ id: item.id, name: item.name })));
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载延迟监测配置失败');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const modalTitle = useMemo(() => (editingTask ? '编辑延迟任务' : '新增延迟任务'), [editingTask]);

  const openCreate = () => {
    setEditingTask(null);
    setForm(defaultFormState);
    setIsOpen(true);
  };

  const openEdit = (task: LatencyTask) => {
    setEditingTask(task);
    setForm({
      name: task.name,
      type: task.type,
      target: task.target,
      intervalSec: task.interval_sec,
      serverIDs: task.server_ids,
    });
    setIsOpen(true);
  };

  const closeModal = () => {
    setIsOpen(false);
    setEditingTask(null);
    setForm(defaultFormState);
  };

  const handleSubmit = async () => {
    setError('');
    const validationError = validateTaskForm(form);
    if (validationError) {
      setError(validationError);
      return;
    }

    try {
      const path = editingTask ? `/admin/latency-tasks/${editingTask.id}` : '/admin/latency-tasks';
      const method = editingTask ? 'PUT' : 'POST';
      await apiRequest(path, {
        method,
        body: JSON.stringify({
          name: form.name.trim(),
          type: form.type,
          target: form.target.trim(),
          interval_sec: form.intervalSec,
          server_ids: form.serverIDs,
        }),
      }, { auth: true });
      closeModal();
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存延迟任务失败');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('确认删除该延迟任务吗？')) {
      return;
    }

    try {
      await apiRequest(`/admin/latency-tasks/${id}`, { method: 'DELETE' }, { auth: true });
      await loadData();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除延迟任务失败');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Latency</h1>
          <p className="text-muted-foreground">按任务配置 TCP / ICMP / HTTP 延迟监测。</p>
        </div>
        <button
          onClick={openCreate}
          className="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow transition-colors hover:bg-primary/90"
        >
          <Plus className="mr-2 h-4 w-4" />
          Add Task
        </button>
      </div>

      {error && (
        <div className="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="rounded-xl border bg-card text-card-foreground shadow">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-muted/50 text-muted-foreground">
              <tr>
                <th className="px-6 py-4 font-medium">名称</th>
                <th className="px-6 py-4 font-medium">类型</th>
                <th className="px-6 py-4 font-medium">目标</th>
                <th className="px-6 py-4 font-medium">服务器</th>
                <th className="px-6 py-4 font-medium">间隔</th>
                <th className="px-6 py-4 font-medium">结果</th>
                <th className="px-6 py-4 text-right font-medium">操作</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {tasks.map((task) => (
                <tr key={task.id} className="align-top transition-colors hover:bg-muted/50">
                  <td className="px-6 py-4 font-medium">{task.name}</td>
                  <td className="px-6 py-4 uppercase">{task.type}</td>
                  <td className="px-6 py-4">{task.target}</td>
                  <td className="px-6 py-4">{task.servers.map((item) => item.name).join(', ') || '-'}</td>
                  <td className="px-6 py-4">{task.interval_sec}s</td>
                  <td className="px-6 py-4">
                    <div className="min-w-[260px] space-y-2">
                      {task.results.length > 0 ? task.results.map((result) => (
                        <div key={`${result.task_id}-${result.server_id}`} className="rounded-lg bg-muted/60 px-3 py-2 text-xs">
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-medium">#{result.server_id}</span>
                            <span>{result.status === 'ok' ? formatLatency(result.latency_ms) : 'Error'}</span>
                          </div>
                          <div className="mt-1 text-muted-foreground">{formatCheckTime(result.checked_at)}</div>
                          {result.error_message ? <div className="mt-1 text-destructive">{result.error_message}</div> : null}
                        </div>
                      )) : <span className="text-muted-foreground">暂无结果</span>}
                    </div>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="flex justify-end gap-2">
                      <button className="rounded-md p-2 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" onClick={() => openEdit(task)}>
                        <Pencil className="h-4 w-4" />
                      </button>
                      <button className="rounded-md p-2 text-destructive transition-colors hover:bg-destructive/10" onClick={() => void handleDelete(task.id)}>
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!isLoading && tasks.length === 0 ? (
            <div className="p-8 text-center text-muted-foreground">暂无延迟任务</div>
          ) : null}
          {isLoading ? (
            <div className="p-8 text-center text-muted-foreground">Loading latency tasks...</div>
          ) : null}
        </div>
      </div>

      {isOpen ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div className="w-full max-w-2xl rounded-xl bg-background shadow-lg">
            <div className="border-b px-6 py-4">
              <h2 className="text-lg font-semibold">{modalTitle}</h2>
            </div>
            <div className="space-y-4 p-6">
              <div className="space-y-2">
                <label className="text-sm font-medium">名称</label>
                <input className="w-full rounded-md border bg-transparent px-3 py-2" value={form.name} onChange={(e) => setForm((prev) => ({ ...prev, name: e.target.value }))} />
              </div>
              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <label className="text-sm font-medium">类型</label>
                  <select className="w-full rounded-md border bg-transparent px-3 py-2" value={form.type} onChange={(e) => setForm((prev) => ({ ...prev, type: e.target.value as TaskFormState['type'] }))}>
                    <option value="tcp">TCP</option>
                    <option value="icmp">ICMP</option>
                    <option value="http">HTTP</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">间隔(秒)</label>
                  <input type="number" min="5" className="w-full rounded-md border bg-transparent px-3 py-2" value={form.intervalSec} onChange={(e) => setForm((prev) => ({ ...prev, intervalSec: Number(e.target.value) }))} />
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">目标</label>
                <input
                  className="w-full rounded-md border bg-transparent px-3 py-2"
                  value={form.target}
                  onChange={(e) => setForm((prev) => ({ ...prev, target: e.target.value }))}
                  placeholder={form.type === 'http' ? 'https://example.com/health' : form.type === 'tcp' ? 'example.com:443' : '8.8.8.8'}
                />
                <p className="text-xs text-muted-foreground">{getTargetHint(form.type)}</p>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium">服务器</label>
                <div className="grid max-h-64 gap-2 overflow-y-auto rounded-md border p-3 md:grid-cols-2">
                  {servers.map((server) => {
                    const checked = form.serverIDs.includes(server.id);
                    return (
                      <label key={server.id} className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={(e) => {
                            setForm((prev) => ({
                              ...prev,
                              serverIDs: e.target.checked
                                ? [...prev.serverIDs, server.id]
                                : prev.serverIDs.filter((id) => id !== server.id),
                            }));
                          }}
                        />
                        {server.name}
                      </label>
                    );
                  })}
                </div>
              </div>
            </div>
            <div className="flex justify-end gap-3 border-t px-6 py-4">
              <button onClick={closeModal} className="rounded-md px-4 py-2 text-sm font-medium text-muted-foreground hover:bg-muted">取消</button>
              <button onClick={() => void handleSubmit()} className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90">保存</button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
