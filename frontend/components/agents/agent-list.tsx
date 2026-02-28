"use client";

import { useState } from "react";
import { toast } from "sonner";
import { useAgents } from "@/hooks/use-agents";
import { useDeployment } from "@/hooks/use-deployment";
import { formatRelativeTime, getStatusColor, cn } from "@/lib/utils";
import { POSITION_LABELS, type Agent, type AgentDeployment } from "@/lib/types";
import { useRouter } from "next/navigation";

interface AgentCardProps {
  agent: Agent;
  onDelete: (id: string) => void;
  onUpdated: () => void;
  isChairman: boolean;
}

function DeployBadge({ status }: { status: AgentDeployment["status"] }) {
  const styles: Record<string, string> = {
    running: "text-green-400 bg-green-500/10 border-green-500/20",
    pending: "text-yellow-400 bg-yellow-500/10 border-yellow-500/20",
    stopped: "text-zinc-400 bg-zinc-500/10 border-zinc-500/20",
    failed:  "text-red-400 bg-red-500/10 border-red-500/20",
  };
  const labels: Record<string, string> = {
    running: "运行中", pending: "部署中", stopped: "已停止", failed: "部署失败",
  };
  return (
    <span className={cn("text-xs px-1.5 py-0.5 rounded border", styles[status] ?? styles.stopped)}>
      {labels[status] ?? status}
    </span>
  );
}

function StatusDot({ agent, deployment }: { agent: Agent; deployment: AgentDeployment | null }) {
  // 部署异常状态优先展示
  if (deployment && deployment.status !== "running") {
    const cfg: Record<string, { dot: string; label: string }> = {
      pending: { dot: "bg-yellow-400", label: "部署中" },
      stopped: { dot: "bg-zinc-400",   label: "已停止" },
      failed:  { dot: "bg-red-400",    label: "部署失败" },
    };
    const s = cfg[deployment.status] ?? cfg.stopped;
    return (
      <div className="flex items-center gap-1.5">
        <span className={cn("w-2 h-2 rounded-full", s.dot)} />
        <span className="text-xs text-zinc-500">{s.label}</span>
      </div>
    );
  }
  return (
    <div className="flex items-center gap-1.5">
      <span className={cn("w-2 h-2 rounded-full", getStatusColor(agent.status))} />
      <span className="text-xs text-zinc-500">{agent.status}</span>
    </div>
  );
}

function AgentCard({ agent, onDelete, onUpdated, isChairman }: AgentCardProps) {
  const router = useRouter();
  const [menuOpen, setMenuOpen] = useState(false);
  const [deployment, setDeployment] = useState<AgentDeployment | null>(null);
  const [loadedDeploy, setLoadedDeploy] = useState(false);
  const { getDeployment, stopDeployment, rebuildDeployment } = useDeployment();

  async function loadDeployment() {
    if (loadedDeploy) return;
    const d = await getDeployment(agent.id);
    setDeployment(d);
    setLoadedDeploy(true);
  }

  async function handleStop() {
    if (!deployment) return;
    await stopDeployment(agent.id);
    setDeployment(prev => prev ? { ...prev, status: "stopped" } : null);
    setMenuOpen(false);
  }

  async function handleRebuild() {
    setMenuOpen(false);
    const result = await rebuildDeployment(agent.id);
    if (result) {
      setDeployment(result.deployment);
      toast.success(`容器已重建，新 API Key: ${result.newApiKey}`, { duration: 15000 });
    } else {
      toast.error("重建失败");
    }
  }

  function handleDelete() {
    setMenuOpen(false);
    toast(`确认删除 Agent "${agent.name}"？此操作不可撤销。`, {
      action: { label: "删除", onClick: () => onDelete(agent.id) },
      cancel: { label: "取消", onClick: () => {} },
    });
  }

  return (
    <div
      className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 hover:border-zinc-700 transition-colors relative"
      onMouseEnter={loadDeployment}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h3 className="font-medium text-zinc-50 truncate">{agent.name}</h3>
          </div>
          <p className="text-zinc-400 text-sm">
            {POSITION_LABELS[agent.position] ?? agent.position}
          </p>
        </div>
        <div className="flex items-center gap-2 ml-2 flex-shrink-0">
          <StatusDot agent={agent} deployment={deployment} />
          {isChairman && !agent.is_human && (
            <div className="relative">
              <button
                onClick={() => setMenuOpen(v => !v)}
                className="p-1 rounded text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800 transition-colors"
              >
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 16 16">
                  <circle cx="8" cy="3" r="1.5" />
                  <circle cx="8" cy="8" r="1.5" />
                  <circle cx="8" cy="13" r="1.5" />
                </svg>
              </button>
              {menuOpen && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setMenuOpen(false)} />
                  <div className="absolute right-0 top-7 z-20 w-36 bg-zinc-800 border border-zinc-700 rounded-lg shadow-xl py-1 text-sm">
                    {deployment?.status === "running" && (
                      <button onClick={handleStop}
                        className="w-full text-left px-3 py-2 text-yellow-400 hover:bg-zinc-700 transition-colors">
                        停止容器
                      </button>
                    )}
                    {(deployment?.status === "stopped" || deployment?.status === "failed") && (
                      <button onClick={handleRebuild}
                        className="w-full text-left px-3 py-2 text-blue-400 hover:bg-zinc-700 transition-colors">
                        重建容器
                      </button>
                    )}
                    <button onClick={() => { setMenuOpen(false); router.push("/prompts"); }}
                      className="w-full text-left px-3 py-2 text-zinc-300 hover:bg-zinc-700 transition-colors">
                      编辑提示词
                    </button>
                    <button onClick={handleDelete}
                      className="w-full text-left px-3 py-2 text-red-400 hover:bg-zinc-700 transition-colors">
                      删除 Agent
                    </button>
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {agent.persona && (
        <p
          onClick={isChairman && !agent.is_human ? () => router.push("/prompts") : undefined}
          className={cn(
            "text-zinc-500 text-xs mt-2 line-clamp-2",
            isChairman && !agent.is_human && "cursor-pointer hover:text-zinc-400 transition-colors"
          )}
        >
          {agent.persona}
        </p>
      )}

      <div className="mt-3 pt-3 border-t border-zinc-800 flex items-center justify-between">
        <span className="text-zinc-600 text-xs font-mono">
          {agent.api_key_prefix ? `${agent.api_key_prefix}…` : "human"}
        </span>
        <span className="text-zinc-600 text-xs">{formatRelativeTime(agent.last_seen_at)}</span>
      </div>

    </div>
  );
}

interface AgentListProps {
  onOpenCreate: () => void;
  isChairman: boolean;
}

export function AgentList({ onOpenCreate, isChairman }: AgentListProps) {
  const { agents, isLoading, error, mutate } = useAgents();

  async function handleDelete(id: string) {
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/agents/${id}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) throw new Error("删除失败");
      toast.success("Agent 已删除");
      mutate();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "删除失败");
    }
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
        <p className="text-red-400 text-sm">加载失败：{error.message}</p>
      </div>
    );
  }

  if (agents.length === 0) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-12 text-center">
        <p className="text-zinc-400">暂无 Agent</p>
        {isChairman && (
          <button onClick={onOpenCreate}
            className="mt-4 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm text-white transition-colors">
            创建第一个 Agent
          </button>
        )}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {agents.map(agent => (
        <AgentCard
          key={agent.id}
          agent={agent}
          onDelete={handleDelete}
          onUpdated={() => mutate()}
          isChairman={isChairman}
        />
      ))}
    </div>
  );
}
