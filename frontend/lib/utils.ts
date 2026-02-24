import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | null): string {
  if (!date) return "—";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(date));
}

export function formatRelativeTime(date: string | null): string {
  if (!date) return "从未";
  const now = Date.now();
  const diff = now - new Date(date).getTime();
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes} 分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} 小时前`;
  const days = Math.floor(hours / 24);
  return `${days} 天前`;
}

export function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    online: "bg-green-500",
    busy: "bg-yellow-500",
    offline: "bg-zinc-400",
    pending: "bg-zinc-400",
    assigned: "bg-blue-500",
    in_progress: "bg-yellow-500",
    done: "bg-green-500",
    failed: "bg-red-500",
    cancelled: "bg-zinc-400",
  };
  return colors[status] ?? "bg-zinc-400";
}

export function getPriorityColor(priority: string): string {
  const colors: Record<string, string> = {
    low: "text-zinc-500",
    medium: "text-blue-500",
    high: "text-orange-500",
    urgent: "text-red-500",
  };
  return colors[priority] ?? "text-zinc-500";
}
