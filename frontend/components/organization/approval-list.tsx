"use client";

import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useApprovals, approveRequest, rejectRequest } from "@/hooks/use-organization";
import { useAgents } from "@/hooks/use-agents";
import { cn, formatDate } from "@/lib/utils";
import type { ApprovalStatus, ApprovalRequestType } from "@/lib/types";

interface ApprovalListProps {
  isChairman: boolean;
}

type FilterStatus = "all" | ApprovalStatus;

const FILTERS: { value: FilterStatus; label: string }[] = [
  { value: "all", label: "全部" },
  { value: "pending", label: "待处理" },
  { value: "approved", label: "已通过" },
  { value: "rejected", label: "已拒绝" },
  { value: "cancelled", label: "已取消" },
];

const STATUS_LABELS: Record<ApprovalStatus, string> = {
  pending: "待处理",
  approved: "已通过",
  rejected: "已拒绝",
  cancelled: "已取消",
};

const STATUS_STYLES: Record<ApprovalStatus, string> = {
  pending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
  approved: "bg-green-500/10 text-green-400 border-green-500/20",
  rejected: "bg-red-500/10 text-red-400 border-red-500/20",
  cancelled: "bg-zinc-500/10 text-zinc-400 border-zinc-500/20",
};

const TYPE_LABELS: Record<ApprovalRequestType, string> = {
  hire: "招聘申请",
  fire: "解雇申请",
  budget_override: "预算超限",
  task_escalation: "任务升级",
  custom: "自定义",
};

function StatusBadge({ status }: { status: ApprovalStatus }) {
  return (
    <span className={cn("inline-flex px-2 py-0.5 rounded border text-xs", STATUS_STYLES[status])}>
      {STATUS_LABELS[status]}
    </span>
  );
}

export function ApprovalList({ isChairman }: ApprovalListProps) {
  const [filter, setFilter] = useState<FilterStatus>("all");
  const [actingId, setActingId] = useState<string | null>(null);
  const [decisionReasons, setDecisionReasons] = useState<Record<string, string>>({});

  const { approvals, isLoading, error, mutate } = useApprovals({
    status: filter === "all" ? undefined : filter,
    limit: 100,
    offset: 0,
  });
  const { agents } = useAgents();

  const requesterNameById = useMemo(
    () => new Map(agents.map((agent) => [agent.id, agent.name])),
    [agents]
  );

  useEffect(() => {
    if (error) toast.error(error.message || "加载审批请求失败");
  }, [error]);

  function requesterLabel(requesterId: string) {
    return requesterNameById.get(requesterId) ?? `${requesterId.slice(0, 8)}…`;
  }

  async function handleDecision(id: string, action: "approve" | "reject") {
    const reason = (decisionReasons[id] ?? "").trim();
    if (!reason) {
      toast.error("请填写 decision_reason");
      return;
    }

    setActingId(id);
    try {
      if (action === "approve") {
        await approveRequest(id, reason);
        toast.success("审批已通过");
      } else {
        await rejectRequest(id, reason);
        toast.success("审批已拒绝");
      }
      setDecisionReasons((prev) => {
        const next = { ...prev };
        delete next[id];
        return next;
      });
      await mutate();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "审批操作失败");
    } finally {
      setActingId(null);
    }
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div className="p-4 border-b border-zinc-800">
        <div className="flex flex-wrap gap-2">
          {FILTERS.map((item) => (
            <button
              key={item.value}
              onClick={() => setFilter(item.value)}
              className={cn(
                "px-3 py-1.5 rounded-md text-xs transition-colors border",
                filter === item.value
                  ? "bg-blue-500/10 text-blue-400 border-blue-500/30"
                  : "bg-zinc-950 text-zinc-400 border-zinc-800 hover:text-zinc-200"
              )}
            >
              {item.label}
            </button>
          ))}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left font-medium px-4 py-3">类型</th>
              <th className="text-left font-medium px-4 py-3">状态</th>
              <th className="text-left font-medium px-4 py-3">发起人</th>
              <th className="text-left font-medium px-4 py-3">原因</th>
              <th className="text-left font-medium px-4 py-3">创建时间</th>
              <th className="text-right font-medium px-4 py-3">操作</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800 animate-pulse">
                  <td className="px-4 py-3"><div className="h-4 w-20 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-5 w-16 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-24 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-48 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-28 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-8 w-40 bg-zinc-800 rounded ml-auto" /></td>
                </tr>
              ))
            ) : approvals.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-zinc-500">
                  暂无审批请求
                </td>
              </tr>
            ) : (
              approvals.map((item) => (
                <tr key={item.id} className="border-t border-zinc-800 hover:bg-zinc-950/50 transition-colors">
                  <td className="px-4 py-3 text-zinc-300">{TYPE_LABELS[item.request_type]}</td>
                  <td className="px-4 py-3"><StatusBadge status={item.status} /></td>
                  <td className="px-4 py-3 text-zinc-400">{requesterLabel(item.requester_id)}</td>
                  <td className="px-4 py-3 text-zinc-400 max-w-[360px] truncate" title={item.reason}>
                    {item.reason}
                  </td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(item.created_at)}</td>
                  <td className="px-4 py-3">
                    {isChairman && item.status === "pending" ? (
                      <div className="flex items-center justify-end gap-2">
                        <input
                          value={decisionReasons[item.id] ?? ""}
                          onChange={(e) => setDecisionReasons((prev) => ({ ...prev, [item.id]: e.target.value }))}
                          className="w-40 bg-zinc-950 border border-zinc-800 rounded px-2 py-1.5 text-xs text-zinc-50 focus:outline-none focus:border-blue-500"
                          placeholder="decision_reason"
                        />
                        <button onClick={() => handleDecision(item.id, "approve")}
                          disabled={actingId === item.id}
                          className="px-2 py-1 rounded text-xs text-green-400 hover:bg-green-500/10 disabled:opacity-50">
                          通过
                        </button>
                        <button onClick={() => handleDecision(item.id, "reject")}
                          disabled={actingId === item.id}
                          className="px-2 py-1 rounded text-xs text-red-400 hover:bg-red-500/10 disabled:opacity-50">
                          拒绝
                        </button>
                      </div>
                    ) : (
                      <div className="text-right text-zinc-500">—</div>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
