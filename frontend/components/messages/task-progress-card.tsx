import { cn, getPriorityColor } from "@/lib/utils";
import { CheckCircle2, XCircle, Clock, Loader2 } from "lucide-react";

interface TaskMeta {
  task_id: string;
  title: string;
  status: string;
  priority: string;
  assignee_id?: string;
  due_at?: string;
  result?: string;
}

interface TaskProgressCardProps {
  meta: TaskMeta;
}

const statusConfig: Record<string, { label: string; icon: React.ElementType; color: string; progress: number }> = {
  pending:     { label: "待分配", icon: Clock,       color: "text-zinc-400",  progress: 0  },
  assigned:    { label: "已分配", icon: Clock,       color: "text-blue-400",  progress: 20 },
  in_progress: { label: "进行中", icon: Loader2,     color: "text-yellow-400", progress: 60 },
  done:        { label: "已完成", icon: CheckCircle2, color: "text-green-400", progress: 100 },
  failed:      { label: "已失败", icon: XCircle,     color: "text-red-400",   progress: 100 },
  cancelled:   { label: "已取消", icon: XCircle,     color: "text-zinc-500",  progress: 0  },
};

export function TaskProgressCard({ meta }: TaskProgressCardProps) {
  const config = statusConfig[meta.status] ?? statusConfig.pending;
  const Icon = config.icon;

  return (
    <div className="my-1 border border-zinc-700 rounded-lg p-3 bg-zinc-900/60 max-w-sm">
      <div className="flex items-start gap-2">
        <Icon className={cn("w-4 h-4 mt-0.5 flex-shrink-0", config.color,
          meta.status === "in_progress" ? "animate-spin" : ""
        )} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="text-sm font-medium text-zinc-100 truncate">{meta.title}</span>
            <span className={cn("text-xs flex-shrink-0", getPriorityColor(meta.priority))}>
              {meta.priority}
            </span>
          </div>
          <div className={cn("text-xs mt-0.5", config.color)}>{config.label}</div>

          {/* 进度条 */}
          <div className="mt-2 h-1.5 bg-zinc-800 rounded-full overflow-hidden">
            <div
              className={cn(
                "h-full rounded-full transition-all duration-500",
                meta.status === "done"   ? "bg-green-500" :
                meta.status === "failed" ? "bg-red-500"   :
                "bg-blue-500"
              )}
              style={{ width: `${config.progress}%` }}
            />
          </div>

          {meta.result && (
            <p className="text-xs text-zinc-400 mt-1.5 line-clamp-2">{meta.result}</p>
          )}
        </div>
      </div>
    </div>
  );
}
