import { useEffect, useState } from 'react';
import { X } from 'lucide-react';
import { DeployCommandModal, type DeployCommands } from './DeployCommandModal';

export interface AddServerFormValues {
  name: string;
  trafficLimit: number;
  trafficResetDay: number;
  expiryTime: string;
}

interface AddServerModalProps {
  isOpen: boolean;
  onClose: () => void;
  mode?: 'create' | 'edit';
  initialValues?: AddServerFormValues | null;
  onSubmit: (values: AddServerFormValues) => Promise<DeployCommands | void>;
}

const defaultFormValues: AddServerFormValues = {
  name: '',
  trafficLimit: 0,
  trafficResetDay: 1,
  expiryTime: '',
};

export function AddServerModal({
  isOpen,
  onClose,
  mode = 'create',
  initialValues = defaultFormValues,
  onSubmit,
}: AddServerModalProps) {
  const [name, setName] = useState('');
  const [trafficLimit, setTrafficLimit] = useState<number>(0);
  const [trafficResetDay, setTrafficResetDay] = useState<number>(1);
  const [expiryTime, setExpiryTime] = useState<string>('');
  const [deployCommands, setDeployCommands] = useState<DeployCommands | null>(null);
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const syncForm = () => {
    setName(initialValues?.name ?? '');
    setTrafficLimit(initialValues?.trafficLimit ? initialValues.trafficLimit / 1024 / 1024 / 1024 : 0);
    setTrafficResetDay(initialValues?.trafficResetDay ?? 1);
    setExpiryTime(initialValues?.expiryTime ? initialValues.expiryTime.slice(0, 10) : '');
    setError('');
    setDeployCommands(null);
    setIsSubmitting(false);
  };

  useEffect(() => {
    if (isOpen) {
      syncForm();
    }
  }, [initialValues, isOpen]);

  if (!isOpen) return null;

  const handleNext = async () => {
    if (!name.trim()) {
      return;
    }

    setError('');
    setIsSubmitting(true);

    try {
      let expiryIso = '';
      if (expiryTime) {
        expiryIso = new Date(expiryTime).toISOString();
      }

      const result = await onSubmit({
        name: name.trim(),
        trafficLimit: trafficLimit * 1024 * 1024 * 1024,
        trafficResetDay,
        expiryTime: expiryIso,
      });

      if (mode === 'create') {
        setDeployCommands(result ?? null);
      } else {
        handleClose();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : mode === 'create' ? '新增服务器失败' : '编辑服务器失败');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleClose = () => {
    syncForm();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-md rounded-xl bg-background shadow-lg overflow-hidden animate-in fade-in zoom-in duration-200">
        <div className="flex items-center justify-between px-6 py-4 border-b">
          <h2 className="text-lg font-semibold">{mode === 'create' ? 'Add New Server' : 'Edit Server'}</h2>
          <button onClick={handleClose} className="p-1 rounded-md hover:bg-muted transition-colors">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Server Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Web-02"
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="rounded-md border border-sky-200 bg-sky-50 px-3 py-2 text-xs text-sky-700 dark:border-sky-900/30 dark:bg-sky-950/20 dark:text-sky-300">
              公网 IP、地区、系统、IPv4 和 IPv6 状态会在 Agent 首次成功上报后自动检测并更新。
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Traffic Limit (GB, 0 for unlimited)</label>
              <input
                type="number"
                min="0"
                value={trafficLimit}
                onChange={(e) => setTrafficLimit(Number(e.target.value))}
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Traffic Reset Day (1-31)</label>
              <input
                type="number"
                min="1"
                max="31"
                value={trafficResetDay}
                onChange={(e) => setTrafficResetDay(Number(e.target.value))}
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Expiry Date (optional)</label>
              <input
                type="date"
                value={expiryTime}
                onChange={(e) => setExpiryTime(e.target.value)}
                className="w-full px-3 py-2 border rounded-md bg-transparent focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
            {error && (
              <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
            <button
              onClick={handleNext}
              disabled={!name.trim() || isSubmitting}
              className="w-full mt-4 py-2 bg-primary text-primary-foreground rounded-md font-medium disabled:opacity-50 hover:bg-primary/90 transition-colors"
            >
              {isSubmitting ? (mode === 'create' ? 'Creating...' : 'Saving...') : mode === 'create' ? 'Create and Get Command' : 'Save Changes'}
            </button>
          </div>
        </div>
      </div>
      <DeployCommandModal
        isOpen={mode === 'create' && deployCommands !== null}
        commands={deployCommands}
        title="一键部署指令"
        onClose={handleClose}
      />
    </div>
  );
}
