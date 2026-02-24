"use client";

import { useEffect, useState } from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import {
  POSITION_LABELS,
  DEPARTMENTS,
  type PromptListResponse,
  type PromptAgentBrief,
} from "@/lib/types";
import { upsertPrompt, deletePrompt, previewPrompt } from "@/hooks/use-prompts";
import type { PromptSelection } from "./prompt-nav";

interface Props {
  selected: PromptSelection | null;
  data: PromptListResponse | null;
  onSaved: () => void;
}

function getTitle(sel: PromptSelection): string {
  switch (sel.type) {
    case "global":
      return "全局提示词";
    case "department":
      return `部门提示词 — ${sel.key}`;
    case "position":
      return `职位提示词 — ${POSITION_LABELS[sel.key] ?? sel.key}`;
    case "agent":
      return `Agent 专属 — ${sel.name}`;
  }
}

function getContent(sel: PromptSelection, data: PromptListResponse | null): string {
  if (!data) return "";
  switch (sel.type) {
    case "global":
      return data.global ?? "";
    case "department":
      return data.departments[sel.key] ?? "";
    case "position":
      return data.positions[sel.key] ?? "";
    case "agent":
      return data.agents?.find((a) => a.id === sel.key)?.persona ?? "";
  }
}

function getDescription(sel: PromptSelection): string {
  switch (sel.type) {
    case "global":
      return "全局提示词影响公司所有 Agent，会出现在每个 Agent 员工手册中。";
    case "department":
      return `此提示词影响「${sel.key}」部门的所有 Agent。`;
    case "position":
      return `此提示词仅影响「${POSITION_LABELS[sel.key] ?? sel.key}」职位的 Agent。`;
    case "agent":
      return "此提示词仅影响该 Agent，优先级最高。";
  }
}

export function PromptEditor({ selected, data, onSaved }: Props) {
  const [content, setContent] = useState("");
  const [saving, setSaving] = useState(false);
  const [previewing, setPreviewing] = useState(false);
  const [previewText, setPreviewText] = useState<string | null>(null);
  const [previewAgentId, setPreviewAgentId] = useState("");

  useEffect(() => {
    if (selected && data) {
      setContent(getContent(selected, data));
      setPreviewText(null);
    }
  }, [selected, data]);

  if (!selected) {
    return (
      <div className="flex-1 flex items-center justify-center text-zinc-600">
        <p>从左侧选择要编辑的提示词层</p>
      </div>
    );
  }

  const title = getTitle(selected);
  const desc = getDescription(selected);

  async function handleSave() {
    if (!selected) return;
    setSaving(true);
    try {
      await upsertPrompt(selected.type, selected.key, content);
      toast.success("提示词已保存");
      onSaved();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (!selected) return;
    setSaving(true);
    try {
      await deletePrompt(selected.type, selected.key);
      setContent("");
      toast.success("提示词已清除");
      onSaved();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "删除失败");
    } finally {
      setSaving(false);
    }
  }

  async function handlePreview() {
    const agentId = previewAgentId || data?.agents?.[0]?.id;
    if (!agentId) {
      toast.error("没有可预览的 Agent");
      return;
    }
    setPreviewing(true);
    try {
      const res = await previewPrompt(agentId);
      setPreviewText(res.prompt);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "预览失败");
    } finally {
      setPreviewing(false);
    }
  }

  const agents = data?.agents ?? [];

  return (
    <div className="flex-1 flex flex-col min-w-0">
      {/* 头部 */}
      <div className="px-6 h-12 flex items-center border-b border-zinc-800 flex-shrink-0">
        <h3 className="text-sm font-semibold text-zinc-200 truncate">{title}</h3>
      </div>

      {/* 编辑区 */}
      <div className="flex-1 overflow-y-auto p-6 space-y-4">
        <p className="text-xs text-zinc-500">{desc}</p>

        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={12}
          placeholder="输入提示词内容..."
          className="w-full p-4 bg-zinc-900 border border-zinc-700 rounded-lg text-zinc-200 text-sm placeholder-zinc-600 focus:outline-none focus:border-blue-500 resize-y font-mono leading-relaxed"
        />

        {/* 操作栏 */}
        <div className="flex items-center gap-3 flex-wrap">
          <button
            onClick={handleSave}
            disabled={saving}
            className="px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-sm font-medium text-white transition-colors disabled:opacity-40"
          >
            {saving ? "保存中..." : "保存"}
          </button>
          <button
            onClick={() => {
              toast("确认清除此层提示词？", {
                action: { label: "清除", onClick: handleDelete },
                cancel: { label: "取消", onClick: () => {} },
              });
            }}
            disabled={saving || !content}
            className="px-4 py-2 rounded-lg text-sm text-zinc-400 hover:text-red-400 hover:bg-zinc-800 transition-colors disabled:opacity-30"
          >
            清除
          </button>

          <div className="flex-1" />

          {/* 预览 */}
          {agents.length > 0 && (
            <div className="flex items-center gap-2">
              <select
                value={previewAgentId || agents[0]?.id || ""}
                onChange={(e) => setPreviewAgentId(e.target.value)}
                className="bg-zinc-800 border border-zinc-700 rounded-md px-2 py-1.5 text-xs text-zinc-300 focus:outline-none focus:border-blue-500"
              >
                {agents.map((a: PromptAgentBrief) => (
                  <option key={a.id} value={a.id}>
                    {a.name || a.id.slice(0, 8)}
                  </option>
                ))}
              </select>
              <button
                onClick={handlePreview}
                disabled={previewing}
                className="px-4 py-1.5 rounded-md text-xs text-zinc-300 bg-zinc-800 border border-zinc-700 hover:border-zinc-600 transition-colors disabled:opacity-40"
              >
                {previewing ? "加载中..." : "预览拼接"}
              </button>
            </div>
          )}
        </div>

        {/* 预览结果 */}
        {previewText !== null && (
          <div className="border border-zinc-700 rounded-lg overflow-hidden">
            <div className="px-4 py-2 bg-zinc-800 border-b border-zinc-700 flex items-center justify-between">
              <span className="text-xs font-medium text-zinc-300">拼接预览</span>
              <button
                onClick={() => setPreviewText(null)}
                className="text-xs text-zinc-500 hover:text-zinc-300"
              >
                关闭
              </button>
            </div>
            <div className="p-4 bg-zinc-900/50 max-h-96 overflow-y-auto">
              <pre className="text-xs text-zinc-300 whitespace-pre-wrap font-mono leading-relaxed">
                {previewText || "(空)"}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
