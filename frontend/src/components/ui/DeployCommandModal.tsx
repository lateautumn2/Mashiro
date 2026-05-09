import { useMemo, useState } from 'react';
import { Check, Copy, X } from 'lucide-react';

export interface DeployCommands {
  linux: string;
  windows: string;
  macos: string;
}

type PlatformKey = keyof DeployCommands;

const PLATFORM_LABELS: Record<PlatformKey, string> = {
  linux: 'Linux',
  windows: 'Windows',
  macos: 'macOS',
};

interface DeployCommandModalProps {
  isOpen: boolean;
  title?: string;
  commands: DeployCommands | null;
  onClose: () => void;
}

export function DeployCommandModal({
  isOpen,
  title = '一键部署指令',
  commands,
  onClose,
}: DeployCommandModalProps) {
  const [platform, setPlatform] = useState<PlatformKey>('linux');
  const [copied, setCopied] = useState(false);

  const currentCommand = useMemo(() => {
    if (!commands) {
      return '';
    }

    return commands[platform];
  }, [commands, platform]);

  if (!isOpen || !commands) {
    return null;
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(currentCommand);
    setCopied(true);
    window.setTimeout(() => setCopied(false), 2000);
  };

  const handleClose = () => {
    setPlatform('linux');
    setCopied(false);
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="w-full max-w-2xl overflow-hidden rounded-xl bg-background shadow-lg animate-in fade-in zoom-in duration-200">
        <div className="flex items-center justify-between border-b px-6 py-4">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button onClick={handleClose} className="rounded-md p-1 transition-colors hover:bg-muted">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="space-y-5 p-6">
          <div className="grid grid-cols-3 rounded-lg bg-muted p-1">
            {(Object.keys(PLATFORM_LABELS) as PlatformKey[]).map((key) => (
              <button
                key={key}
                onClick={() => setPlatform(key)}
                className={`rounded-md px-4 py-2 text-sm font-medium transition-colors ${platform === key ? 'bg-background shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
              >
                {PLATFORM_LABELS[key]}
              </button>
            ))}
          </div>

          <div className="space-y-2">
            <div className="text-sm font-medium">指令</div>
            <div className="relative">
              <pre className="overflow-x-auto rounded-lg border bg-muted/50 p-4 text-sm">
                <code>{currentCommand}</code>
              </pre>
              <button
                onClick={() => void handleCopy()}
                className="absolute right-2 top-2 rounded-md border bg-background p-1.5 shadow-sm transition-colors hover:bg-muted"
              >
                {copied ? <Check className="h-4 w-4 text-emerald-500" /> : <Copy className="h-4 w-4" />}
              </button>
            </div>
          </div>

          <button
            onClick={handleClose}
            className="w-full rounded-md bg-primary py-2 font-medium text-primary-foreground transition-colors hover:bg-primary/90"
          >
            完成
          </button>
        </div>
      </div>
    </div>
  );
}
