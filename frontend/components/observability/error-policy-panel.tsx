"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { createErrorPolicy, useErrorPolicies } from "@/hooks/use-observability";
import { formatDate } from "@/lib/utils";
import type { ErrorAlertScopeType } from "@/lib/types";

interface ErrorPolicyFormState {
  scopeType: ErrorAlertScopeType;
  scopeId: string;
  windowMinutes: string;
  minRequests: string;
  errorRateThresholdPercent: string;
  cooldownMinutes: string;
}

const DEFAULT_FORM: ErrorPolicyFormState = {
  scopeType: "company",
  scopeId: "",
  windowMinutes: "5",
  minRequests: "20",
  errorRateThresholdPercent: "5",
  cooldownMinutes: "10",
};

const SCOPE_LABELS: Record<ErrorAlertScopeType, string> = {
  company: "公司",
  provider: "Provider",
  model: "模型",
  agent: "Agent",
};

export function ErrorPolicyPanel() {
  const [form, setForm] = useState<ErrorPolicyFormState>(DEFAULT_FORM);
  const [saving, setSaving] = useState(false);
  const { policies, isLoading, error, mutate } = useErrorPolicies();

  useEffect(() => {
    if (error) toast.error(error instanceof Error ? error.message : "操作失败");
  }, [error]);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    const thresholdPercent = Number(form.errorRateThresholdPercent);
    if (!Number.isFinite(thresholdPercent) || thresholdPercent <= 0 || thresholdPercent > 100) {
      toast.error("错误率阈值应在 0-100 之间");
      return;
    }

    setSaving(true);
    try {
      await createErrorPolicy({
        scopeType: form.scopeType,
        scopeId: form.scopeId.trim() || undefined,
        windowMinutes: Number(form.windowMinutes),
        minRequests: Number(form.minRequests),
        errorRateThreshold: thresholdPercent / 100,
        cooldownMinutes: Number(form.cooldownMinutes),
      });
      toast.success("错误策略已创建");
      setForm(DEFAULT_FORM);
      await mutate();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <form onSubmit={handleCreate} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-3">
        <h3 className="text-zinc-50 font-medium text-sm">创建错误告警策略</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          <select
            value={form.scopeType}
            onChange={(e) => setForm((s) => ({ ...s, scopeType: e.target.value as ErrorAlertScopeType }))}
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
          >
            <option value="company">公司级</option>
            <option value="provider">Provider 级</option>
            <option value="model">模型级</option>
            <option value="agent">Agent 级</option>
          </select>
          <input
            value={form.scopeId}
            onChange={(e) => setForm((s) => ({ ...s, scopeId: e.target.value }))}
            placeholder="Scope ID（可选）"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min={1}
            value={form.windowMinutes}
            onChange={(e) => setForm((s) => ({ ...s, windowMinutes: e.target.value }))}
            placeholder="窗口分钟数"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min={1}
            value={form.minRequests}
            onChange={(e) => setForm((s) => ({ ...s, minRequests: e.target.value }))}
            placeholder="最小请求数"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min={0}
            max={100}
            step="0.1"
            value={form.errorRateThresholdPercent}
            onChange={(e) => setForm((s) => ({ ...s, errorRateThresholdPercent: e.target.value }))}
            placeholder="错误率阈值(%)"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
          <input
            type="number"
            min={1}
            value={form.cooldownMinutes}
            onChange={(e) => setForm((s) => ({ ...s, cooldownMinutes: e.target.value }))}
            placeholder="冷却分钟数"
            className="bg-zinc-950 border border-zinc-800 rounded px-3 py-2 text-sm text-zinc-50 placeholder-zinc-500 focus:outline-none focus:border-blue-500"
          />
        </div>
        <button
          type="submit"
          disabled={saving}
          className="px-4 py-2 rounded bg-blue-600 hover:bg-blue-700 text-white text-sm disabled:opacity-50"
        >
          {saving ? "创建中…" : "创建错误策略"}
        </button>
      </form>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left px-4 py-3 font-medium">范围</th>
              <th className="text-left px-4 py-3 font-medium">窗口</th>
              <th className="text-left px-4 py-3 font-medium">最小请求</th>
              <th className="text-left px-4 py-3 font-medium">错误率阈值</th>
              <th className="text-left px-4 py-3 font-medium">冷却时间</th>
              <th className="text-left px-4 py-3 font-medium">创建时间</th>
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800">
                  <td className="px-4 py-3" colSpan={6}><div className="animate-pulse h-4 bg-zinc-800 rounded" /></td>
                </tr>
              ))
            ) : policies.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-zinc-500">暂无错误策略</td>
              </tr>
            ) : (
              policies.map((policy) => (
                <tr key={policy.id} className="border-t border-zinc-800 hover:bg-zinc-950/50">
                  <td className="px-4 py-3 text-zinc-300">{SCOPE_LABELS[policy.scopeType]} {policy.scopeId ? `(${policy.scopeId.slice(0, 8)}…)` : ""}</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.windowMinutes} 分钟</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.minRequests}</td>
                  <td className="px-4 py-3 text-zinc-300">{(policy.errorRateThreshold * 100).toFixed(2)}%</td>
                  <td className="px-4 py-3 text-zinc-400">{policy.cooldownMinutes} 分钟</td>
                  <td className="px-4 py-3 text-zinc-500">{formatDate(policy.createdAt)}</td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
