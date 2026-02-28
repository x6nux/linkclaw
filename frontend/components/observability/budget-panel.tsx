"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import {
  createBudgetPolicy,
  patchBudgetAlert,
  updateBudgetPolicy,
  useBudgetAlerts,
  useBudgetPolicies,
} from "@/hooks/use-observability";
import { formatDate } from "@/lib/utils";
import type {
  BudgetAlertLevel,
  BudgetAlertStatus,
  BudgetPeriod,
  BudgetScopeType,
  LLMBudgetPolicy,
} from "@/lib/types";

interface BudgetFormState {
  scopeType: BudgetScopeType;
  scopeId: string;
  period: BudgetPeriod;
  budgetUsd: string;
  warnRatio: string;
  criticalRatio: string;
  hardLimitEnabled: boolean;
  isActive: boolean;
}

const DEFAULT_FORM: BudgetFormState = {
  scopeType: "company",
  scopeId: "",
  period: "monthly",
  budgetUsd: "10",
  warnRatio: "0.8",
  criticalRatio: "1",
  hardLimitEnabled: false,
  isActive: true,
};

const SCOPE_LABELS: Record<BudgetScopeType, string> = {
  company: "公司",
  agent: "Agent",
  provider: "Provider",
};

const PERIOD_LABELS: Record<BudgetPeriod, string> = {
  daily: "每日",
  weekly: "每周",
  monthly: "每月",
};

const LEVEL_STYLES: Record<BudgetAlertLevel, string> = {
  warn: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  critical: "bg-red-500/10 text-red-400 border-red-500/20",
  blocked: "bg-purple-500/10 text-purple-400 border-purple-500/20",
};

const LEVEL_LABELS: Record<BudgetAlertLevel, string> = {
  warn: "预警",
  critical: "严重",
  blocked: "阻断",
};

const ALERT_STATUS_LABELS: Record<BudgetAlertStatus, string> = {
  open: "未处理",
  acked: "已确认",
  resolved: "已解决",
};

function formatUsd(microdollars: number) {
  return `$${(microdollars / 1_000_000).toFixed(4)}`;
}

export function BudgetPanel() {
  const [form, setForm] = useState<BudgetFormState>(DEFAULT_FORM);
  const [saving, setSaving] = useState(false);
  const [updatingPolicyId, setUpdatingPolicyId] = useState<string | null>(null);
  const [updatingAlertId, setUpdatingAlertId] = useState<string | null>(null);
  const [alertStatus, setAlertStatus] = useState<Record<string, BudgetAlertStatus>>({});

  const { policies, isLoading: policiesLoading, error: policiesError, mutate: mutatePolicies } = useBudgetPolicies();
  const { alerts, isLoading: alertsLoading, error: alertsError, mutate: mutateAlerts } = useBudgetAlerts({
    limit: 20,
    offset: 0,
  });

  useEffect(() => {
    if (policiesError) toast.error(policiesError instanceof Error ? policiesError.message : "操作失败");
  }, [policiesError]);

  useEffect(() => {
    if (alertsError) toast.error(alertsError instanceof Error ? alertsError.message : "操作失败");
  }, [alertsError]);

  async function handleCreatePolicy(e: React.FormEvent) {
    e.preventDefault();
    const budgetMicrodollars = Math.round(Number(form.budgetUsd) * 1_000_000);
    const warnRatio = Number(form.warnRatio);
    const criticalRatio = Number(form.criticalRatio);

    if (!Number.isFinite(budgetMicrodollars) || budgetMicrodollars <= 0) {
      toast.error("预算金额必须大于 0");
      return;
    }
    if (!Number.isFinite(warnRatio) || !Number.isFinite(criticalRatio) || criticalRatio < warnRatio) {
      toast.error("阈值设置不合法");
      return;
    }

    setSaving(true);
    try {
      await createBudgetPolicy({
        scopeType: form.scopeType,
        scopeId: form.scopeId.trim() || undefined,
        period: form.period,
        budgetMicrodollars,
        warnRatio,
        criticalRatio,
        hardLimitEnabled: form.hardLimitEnabled,
        isActive: form.isActive,
      });
      toast.success("预算策略已创建");
      setForm(DEFAULT_FORM);
      await mutatePolicies();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleTogglePolicy(policy: LLMBudgetPolicy) {
    setUpdatingPolicyId(policy.id);
    try {
      await updateBudgetPolicy(policy.id, {
        budgetMicrodollars: policy.budgetMicrodollars,
        warnRatio: policy.warnRatio,
        criticalRatio: policy.criticalRatio,
        hardLimitEnabled: policy.hardLimitEnabled,
        isActive: !policy.isActive,
      });
      toast.success(policy.isActive ? "策略已停用" : "策略已启用");
      await mutatePolicies();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setUpdatingPolicyId(null);
    }
  }

  async function handlePatchAlert(id: string, currentStatus: BudgetAlertStatus) {
    const nextStatus = alertStatus[id] ?? currentStatus;
    if (nextStatus === currentStatus) return;

    setUpdatingAlertId(id);
    try {
      await patchBudgetAlert(id, nextStatus);
      toast.success("告警状态已更新");
      await mutateAlerts();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setUpdatingAlertId(null);
    }
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleCreatePolicy} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h3 className="text-zinc-50 font-medium text-sm">创建预算策略</h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
          <select
            value={form.scopeType}
            onChange={(e) => setForm((s) => ({ ...s, scopeType: e.target.value as BudgetScopeType }))}
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
          >
            <option value="company">公司级</option>
            <option value="agent">Agent 级</option>
            <option value="provider">Provider 级</option>
          </select>
          <input
            value={form.scopeId}
            onChange={(e) => setForm((s) => ({ ...s, scopeId: e.target.value }))}
            placeholder="Scope ID（可选）"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <select
            value={form.period}
            onChange={(e) => setForm((s) => ({ ...s, period: e.target.value as BudgetPeriod }))}
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
          >
            <option value="daily">每日</option>
            <option value="weekly">每周</option>
            <option value="monthly">每月</option>
          </select>
          <input
            type="number"
            min="0"
            step="0.0001"
            value={form.budgetUsd}
            onChange={(e) => setForm((s) => ({ ...s, budgetUsd: e.target.value }))}
            placeholder="预算（USD）"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min="0"
            step="0.01"
            value={form.warnRatio}
            onChange={(e) => setForm((s) => ({ ...s, warnRatio: e.target.value }))}
            placeholder="Warn Ratio"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min="0"
            step="0.01"
            value={form.criticalRatio}
            onChange={(e) => setForm((s) => ({ ...s, criticalRatio: e.target.value }))}
            placeholder="Critical Ratio"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <label className="flex items-center gap-2 text-sm text-zinc-300 px-1">
            <input
              type="checkbox"
              checked={form.hardLimitEnabled}
              onChange={(e) => setForm((s) => ({ ...s, hardLimitEnabled: e.target.checked }))}
              className="accent-blue-500"
            />
            启用硬限制
          </label>
          <label className="flex items-center gap-2 text-sm text-zinc-300 px-1">
            <input
              type="checkbox"
              checked={form.isActive}
              onChange={(e) => setForm((s) => ({ ...s, isActive: e.target.checked }))}
              className="accent-blue-500"
            />
            启用策略
          </label>
        </div>
        <button
          type="submit"
          disabled={saving}
          className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm disabled:opacity-50"
        >
          {saving ? "创建中…" : "创建预算策略"}
        </button>
      </form>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left px-4 py-3 font-medium">范围</th>
              <th className="text-left px-4 py-3 font-medium">周期</th>
              <th className="text-left px-4 py-3 font-medium">预算</th>
              <th className="text-left px-4 py-3 font-medium">预警</th>
              <th className="text-left px-4 py-3 font-medium">硬限制</th>
              <th className="text-left px-4 py-3 font-medium">状态</th>
              <th className="text-left px-4 py-3 font-medium">创建时间</th>
              <th className="text-right px-4 py-3 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {policiesLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800">
                  <td className="px-4 py-3" colSpan={8}><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                </tr>
              ))
            ) : policies.length === 0 ? (
              <tr><td colSpan={8} className="px-4 py-8 text-center text-zinc-500">暂无预算策略</td></tr>
            ) : (
              policies.map((policy) => (
                <tr key={policy.id} className="border-t border-zinc-800 hover:bg-zinc-950/50">
                  <td className="px-4 py-3 text-zinc-300">{SCOPE_LABELS[policy.scopeType]} {policy.scopeId ? `(${policy.scopeId.slice(0, 8)}…)` : ""}</td>
                  <td className="px-4 py-3 text-zinc-400">{PERIOD_LABELS[policy.period]}</td>
                  <td className="px-4 py-3 text-zinc-300 font-mono">{formatUsd(policy.budgetMicrodollars)}</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.warnRatio} / {policy.criticalRatio}</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.hardLimitEnabled ? "开启" : "关闭"}</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.isActive ? "启用" : "停用"}</td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(policy.createdAt)}</td>
                  <td className="px-4 py-3 text-right">
                    <button
                      onClick={() => handleTogglePolicy(policy)}
                      disabled={updatingPolicyId === policy.id}
                      className="px-2 py-1 rounded text-xs text-zinc-300 hover:bg-zinc-800 disabled:opacity-50"
                    >
                      {policy.isActive ? "停用" : "启用"}
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left px-4 py-3 font-medium">级别</th>
              <th className="text-left px-4 py-3 font-medium">状态</th>
              <th className="text-left px-4 py-3 font-medium">当前成本</th>
              <th className="text-left px-4 py-3 font-medium">周期</th>
              <th className="text-left px-4 py-3 font-medium">创建时间</th>
              <th className="text-right px-4 py-3 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            {alertsLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800">
                  <td className="px-4 py-3" colSpan={6}><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                </tr>
              ))
            ) : alerts.length === 0 ? (
              <tr><td colSpan={6} className="px-4 py-8 text-center text-zinc-500">暂无预算告警</td></tr>
            ) : (
              alerts.map((alert) => (
                <tr key={alert.id} className="border-t border-zinc-800 hover:bg-zinc-950/50">
                  <td className="px-4 py-3">
                    <span className={`inline-flex px-2 py-0.5 rounded border text-xs ${LEVEL_STYLES[alert.level]}`}>
                      {LEVEL_LABELS[alert.level]}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-zinc-400">{ALERT_STATUS_LABELS[alert.status]}</td>
                  <td className="px-4 py-3 text-zinc-300 font-mono">{formatUsd(alert.currentCostMicrodollars)}</td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(alert.periodStart)} - {formatDate(alert.periodEnd)}</td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(alert.createdAt)}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end gap-2">
                      <select
                        value={alertStatus[alert.id] ?? alert.status}
                        onChange={(e) => setAlertStatus((s) => ({ ...s, [alert.id]: e.target.value as BudgetAlertStatus }))}
                        className="bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs text-zinc-50"
                      >
                        <option value="open">未处理</option>
                        <option value="acked">已确认</option>
                        <option value="resolved">已解决</option>
                      </select>
                      <button
                        onClick={() => handlePatchAlert(alert.id, alert.status)}
                        disabled={updatingAlertId === alert.id}
                        className="px-2 py-1 rounded text-xs text-zinc-300 hover:bg-zinc-800 disabled:opacity-50"
                      >
                        更新
                      </button>
                    </div>
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
