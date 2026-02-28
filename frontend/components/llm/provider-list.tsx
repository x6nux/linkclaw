"use client";

import { useState, useRef, KeyboardEvent } from "react";
import useSWR, { mutate } from "swr";
import { api } from "@/lib/api";
import { LLMProvider, ProviderType } from "@/lib/types";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Plus, Pencil, Trash2, CheckCircle2, AlertTriangle, XCircle,
  RefreshCw, X,
} from "lucide-react";

// ===== 状态图标 =====

function StatusBadge({ status }: { status: LLMProvider["status"] }) {
  if (status === "healthy")
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-transparent bg-green-500/15 px-2 py-0.5 text-xs font-semibold text-green-400">
        <CheckCircle2 className="w-3 h-3" />健康
      </span>
    );
  if (status === "degraded")
    return (
      <span className="inline-flex items-center gap-1 rounded-full border border-transparent bg-amber-500/15 px-2 py-0.5 text-xs font-semibold text-amber-400">
        <AlertTriangle className="w-3 h-3" />降级
      </span>
    );
  return (
    <span className="inline-flex items-center gap-1 rounded-full border border-transparent bg-red-500/15 px-2 py-0.5 text-xs font-semibold text-red-400">
      <XCircle className="w-3 h-3" />停用
    </span>
  );
}

// ===== 多模型 Tag 输入 =====

function ModelsInput({
  models,
  suggestions,
  onChange,
}: {
  models: string[];
  suggestions: string[];
  onChange: (models: string[]) => void;
}) {
  const [input, setInput] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  function addModel(val: string) {
    const trimmed = val.trim();
    if (trimmed && !models.includes(trimmed)) {
      onChange([...models, trimmed]);
    }
    setInput("");
  }

  function removeModel(m: string) {
    onChange(models.filter((x) => x !== m));
  }

  function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
    if (e.key === "Enter" || e.key === ",") {
      e.preventDefault();
      addModel(input);
    } else if (e.key === "Backspace" && input === "" && models.length > 0) {
      onChange(models.slice(0, -1));
    }
  }

  return (
    <div
      className="flex flex-wrap gap-1.5 items-center min-h-[2.25rem] w-full bg-zinc-800 border border-zinc-700 rounded px-2 py-1.5 focus-within:border-blue-500 cursor-text"
      onClick={() => inputRef.current?.focus()}
    >
      {models.map((m) => (
        <span
          key={m}
          className="inline-flex items-center gap-1 bg-zinc-700 text-zinc-200 text-xs rounded px-1.5 py-0.5"
        >
          {m}
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); removeModel(m); }}
            className="text-zinc-400 hover:text-zinc-100"
          >
            <X className="w-2.5 h-2.5" />
          </button>
        </span>
      ))}
      <input
        ref={inputRef}
        list="model-suggestions"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        onBlur={() => { if (input.trim()) addModel(input); }}
        className="flex-1 min-w-[8rem] bg-transparent text-sm text-zinc-50 focus:outline-none placeholder-zinc-500"
        placeholder={models.length === 0 ? "输入模型名，回车确认" : "继续添加..."}
      />
      <datalist id="model-suggestions">
        {suggestions.filter((s) => !models.includes(s)).sort().map((m) => (
          <option key={m} value={m} />
        ))}
      </datalist>
    </div>
  );
}

// ===== 表单对话框 =====

interface ProviderFormData {
  name: string;
  type: ProviderType;
  base_url: string;
  api_key: string;
  models: string[];
  weight: number;
  is_active: boolean;
  max_rpm: string;
}

const DEFAULT_BASE_URLS: Record<ProviderType, string> = {
  anthropic: "https://api.anthropic.com",
  openai: "https://api.openai.com",
};

function ProviderForm({
  initial,
  models,
  onSave,
  onCancel,
}: {
  initial?: Partial<LLMProvider>;
  models: Record<ProviderType, string[]>;
  onSave: (data: ProviderFormData) => Promise<void>;
  onCancel: () => void;
}) {
  const [form, setForm] = useState<ProviderFormData>({
    name: initial?.name ?? "",
    type: initial?.type ?? "anthropic",
    base_url: initial?.base_url ?? DEFAULT_BASE_URLS["anthropic"],
    api_key: "",
    models: initial?.models ?? [],
    weight: initial?.weight ?? 100,
    is_active: initial?.is_active ?? true,
    max_rpm: initial?.max_rpm?.toString() ?? "",
  });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const modelSuggestions = models[form.type] ?? [];

  function handleTypeChange(t: ProviderType) {
    setForm((f) => ({
      ...f,
      type: t,
      base_url: DEFAULT_BASE_URLS[t],
      models: [],
    }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (form.models.length === 0) {
      setError("请至少添加一个模型");
      return;
    }
    setSaving(true);
    setError("");
    try {
      await onSave(form);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "保存失败");
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <form
        onSubmit={handleSubmit}
        className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 w-full max-w-md space-y-4"
      >
        <h3 className="text-zinc-50 font-semibold">
          {initial?.id ? "编辑 Provider" : "添加 Provider"}
        </h3>

        <div className="grid grid-cols-2 gap-3">
          <div className="col-span-2">
            <label className="text-xs text-zinc-400 mb-1 block">名称</label>
            <input
              required
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.name}
              onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
              placeholder="如：GLM 主力"
            />
          </div>

          <div>
            <label className="text-xs text-zinc-400 mb-1 block">类型</label>
            <select
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.type}
              onChange={(e) => handleTypeChange(e.target.value as ProviderType)}
            >
              <option value="anthropic">Anthropic</option>
              <option value="openai">OpenAI</option>
            </select>
          </div>

          <div>
            <label className="text-xs text-zinc-400 mb-1 block">权重</label>
            <input
              type="number"
              min={1}
              max={1000}
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.weight}
              onChange={(e) => setForm((f) => ({ ...f, weight: Number(e.target.value) }))}
            />
          </div>

          <div className="col-span-2">
            <label className="text-xs text-zinc-400 mb-1 block">
              模型 <span className="text-zinc-600">（回车添加多个）</span>
            </label>
            <ModelsInput
              models={form.models}
              suggestions={modelSuggestions}
              onChange={(ms) => setForm((f) => ({ ...f, models: ms }))}
            />
          </div>

          <div className="col-span-2">
            <label className="text-xs text-zinc-400 mb-1 block">Base URL</label>
            <input
              required
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.base_url}
              onChange={(e) => setForm((f) => ({ ...f, base_url: e.target.value }))}
            />
          </div>

          <div className="col-span-2">
            <label className="text-xs text-zinc-400 mb-1 block">
              API Key {initial?.id && <span className="text-zinc-500">（留空保持不变）</span>}
            </label>
            <input
              type="password"
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.api_key}
              onChange={(e) => setForm((f) => ({ ...f, api_key: e.target.value }))}
              placeholder={initial?.id ? "不修改则留空" : "sk-..."}
              required={!initial?.id}
            />
          </div>

          <div>
            <label className="text-xs text-zinc-400 mb-1 block">最大 RPM（可选）</label>
            <input
              type="number"
              min={1}
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              value={form.max_rpm}
              onChange={(e) => setForm((f) => ({ ...f, max_rpm: e.target.value }))}
              placeholder="不限制"
            />
          </div>

          <div className="flex items-center gap-2 self-end pb-2">
            <input
              type="checkbox"
              id="is_active"
              checked={form.is_active}
              onChange={(e) => setForm((f) => ({ ...f, is_active: e.target.checked }))}
              className="accent-blue-500"
            />
            <label htmlFor="is_active" className="text-sm text-zinc-300">启用</label>
          </div>
        </div>

        {error && <p className="text-red-400 text-xs">{error}</p>}

        <div className="flex gap-2 justify-end pt-2">
          <Button type="button" variant="ghost" size="sm" onClick={onCancel}>取消</Button>
          <Button type="submit" size="sm" disabled={saving}>
            {saving ? <RefreshCw className="w-3 h-3 animate-spin mr-1" /> : null}
            保存
          </Button>
        </div>
      </form>
    </div>
  );
}

// ===== 主组件 =====

export function ProviderList({ models }: { models: Record<ProviderType, string[]> }) {
  const fetcher = (url: string) => api.get<{ data: LLMProvider[]; total: number }>(url);
  const { data, error } = useSWR("/api/v1/llm/providers", fetcher);
  const [editing, setEditing] = useState<Partial<LLMProvider> | null>(null);

  async function handleSave(form: ProviderFormData) {
    const payload = {
      ...form,
      weight: form.weight,
      max_rpm: form.max_rpm ? Number(form.max_rpm) : undefined,
    };
    if (editing?.id) {
      await api.put(`/api/v1/llm/providers/${editing.id}`, payload);
    } else {
      await api.post("/api/v1/llm/providers", payload);
    }
    mutate("/api/v1/llm/providers");
    mutate("/api/v1/llm/stats");
    setEditing(null);
  }

  async function handleDelete(id: string) {
    if (!confirm("确认删除此 Provider？")) return;
    await api.delete(`/api/v1/llm/providers/${id}`);
    mutate("/api/v1/llm/providers");
    mutate("/api/v1/llm/stats");
  }

  if (error) return <p className="text-red-400 text-sm">加载失败</p>;

  const providers = data?.data ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-zinc-50 font-semibold">Provider 配置</h2>
        <Button size="sm" onClick={() => setEditing({})}>
          <Plus className="w-3 h-3 mr-1" /> 添加
        </Button>
      </div>

      {providers.length === 0 ? (
        <div className="text-center py-12 text-zinc-500 text-sm border border-dashed border-zinc-800 rounded-lg">
          还没有配置任何 Provider，点击「添加」开始
        </div>
      ) : (
        <div className="space-y-2">
          {providers.map((p) => (
            <div
              key={p.id}
              className="flex items-center gap-4 p-4 bg-zinc-900 border border-zinc-800 rounded-lg"
            >
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-zinc-50 text-sm font-medium truncate">{p.name}</span>
                  <Badge variant="outline" className="text-xs capitalize">{p.type}</Badge>
                  <StatusBadge status={p.status} />
                  {!p.is_active && (
                    <Badge variant="secondary" className="text-xs">已禁用</Badge>
                  )}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs text-zinc-500">
                  {p.models.length > 0 ? (
                    p.models.map((m) => (
                      <span key={m} className="bg-zinc-800 text-zinc-300 px-1.5 py-0.5 rounded font-mono">
                        {m}
                      </span>
                    ))
                  ) : (
                    <span className="text-zinc-600 italic">暂无模型</span>
                  )}
                  <span className="text-zinc-600">权重 {p.weight}</span>
                  <span className="font-mono text-zinc-600">{p.api_key_prefix}</span>
                  {p.error_count > 0 && (
                    <span className="text-amber-500">错误 {p.error_count} 次</span>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-1">
                <button
                  onClick={() => setEditing(p)}
                  className="p-1.5 rounded text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 transition-colors"
                >
                  <Pencil className="w-3.5 h-3.5" />
                </button>
                <button
                  onClick={() => handleDelete(p.id)}
                  className="p-1.5 rounded text-zinc-400 hover:text-red-400 hover:bg-zinc-800 transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {editing !== null && (
        <ProviderForm
          initial={editing}
          models={models}
          onSave={handleSave}
          onCancel={() => setEditing(null)}
        />
      )}
    </div>
  );
}
