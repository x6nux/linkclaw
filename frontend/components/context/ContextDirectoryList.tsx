"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { ContextDirectory } from "@/lib/types";
import { cn } from "@/lib/utils";

interface Props {
  onEdit: (dir: ContextDirectory) => void;
  onRefresh: () => void;
}

export function ContextDirectoryList({ onEdit, onRefresh }: Props) {
  const t = useTranslations("context");
  const [directories, setDirectories] = useState<ContextDirectory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadDirectories();
  }, []);

  async function loadDirectories() {
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch("/api/v1/context/directories", {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error(t("errors.networkError"));
      const data = await res.json();
      setDirectories(data.data || []);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : t("errors.networkError"));
    } finally {
      setIsLoading(false);
    }
  }

  async function toggleActive(id: string, current: boolean) {
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/context/directories/${id}/toggle`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ is_active: !current }),
      });
      if (!res.ok) throw new Error(t("errors.operationFailed"));
      toast.success(!current ? t("activateSuccess") : t("deactivateSuccess"));
      loadDirectories();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("errors.operationFailed"));
    }
  }

  async function handleDelete(id: string) {
    toast(t("confirmDelete"), {
      action: {
        label: t("delete"),
        onClick: async () => {
          try {
            const token = localStorage.getItem("lc_token");
            const res = await fetch(`/api/v1/context/directories/${id}`, {
              method: "DELETE",
              headers: { Authorization: `Bearer ${token}` },
            });
            if (!res.ok) throw new Error(t("errors.deleteFailed"));
            toast.success(t("deleteSuccess"));
            loadDirectories();
          } catch (e) {
            toast.error(e instanceof Error ? e.message : t("errors.deleteFailed"));
          }
        }
      },
      cancel: { label: t("common.cancel"), onClick: () => {} },
    });
  }

  function formatTime(dateStr?: string) {
    if (!dateStr) return t("notIndexed");
    const date = new Date(dateStr);
    const now = Date.now();
    const diff = now - date.getTime();
    const minutes = Math.floor(diff / 60000);
    const hours = Math.floor(diff / 3600000);
    const days = Math.floor(diff / 86400000);

    if (minutes < 1) return t("time.justNow");
    if (minutes < 60) return t("time.minutesAgo", { values: { count: minutes } as never });
    if (hours < 24) return t("time.hoursAgo", { values: { count: hours } as never });
    return t("time.daysAgo", { values: { count: days } as never });
  }

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 animate-pulse">
            <div className="h-4 bg-zinc-800 rounded w-3/4 mb-2" />
            <div className="h-3 bg-zinc-800 rounded w-1/2" />
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-center">
        <p className="text-red-400 text-sm">{error}</p>
        <button
          onClick={loadDirectories}
          className="mt-4 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm text-white transition-colors"
        >
          {t("common.retry")}
        </button>
      </div>
    );
  }

  if (directories.length === 0) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
        <p className="text-zinc-400">{t("noDirectories")}</p>
        <p className="text-zinc-500 text-sm mt-2">{t("noDirectoriesHint")}</p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {directories.map((dir) => (
        <div
          key={dir.id}
          className={cn(
            "bg-zinc-900 border rounded-lg p-4 hover:border-zinc-700 transition-colors relative",
            dir.is_active ? "border-zinc-800" : "border-zinc-800/50 opacity-75"
          )}
        >
          <div className="flex items-start justify-between mb-2">
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="font-medium text-zinc-50 truncate">{dir.name}</h3>
                {dir.is_active && (
                  <span className="text-xs px-1.5 py-0.5 rounded bg-green-500/10 text-green-400 border border-green-500/20">
                    {t("active")}
                  </span>
                )}
              </div>
              <p className="text-zinc-400 text-sm font-mono truncate mt-1">{dir.path}</p>
            </div>
          </div>

          {dir.description && (
            <p className="text-zinc-500 text-xs mt-2 line-clamp-2">{dir.description}</p>
          )}

          <div className="mt-3 pt-3 border-t border-zinc-800 flex items-center justify-between text-xs text-zinc-500">
            <span>{t("files", { values: { count: dir.file_count } as never })}</span>
            <span>{formatTime(dir.last_indexed_at)}</span>
          </div>

          <div className="mt-3 flex items-center justify-between">
            <button
              onClick={() => toggleActive(dir.id, dir.is_active)}
              className={cn(
                "text-xs px-2 py-1 rounded transition-colors",
                dir.is_active
                  ? "text-yellow-400 hover:bg-yellow-500/10"
                  : "text-green-400 hover:bg-green-500/10"
              )}
            >
              {dir.is_active ? t("deactivate") : t("activate")}
            </button>
            <div className="flex items-center gap-1">
              <button
                onClick={() => onEdit(dir)}
                className="text-xs px-2 py-1 rounded text-blue-400 hover:bg-blue-500/10 transition-colors"
              >
                {t("edit")}
              </button>
              <button
                onClick={() => handleDelete(dir.id)}
                className="text-xs px-2 py-1 rounded text-red-400 hover:bg-red-500/10 transition-colors"
              >
                {t("delete")}
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
