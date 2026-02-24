"use client";

import { useState, useEffect } from "react";
import { toast } from "sonner";
import { Save, Trash2, Search } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Memory,
  Agent,
  MemoryImportance,
  IMPORTANCE_LABELS,
  IMPORTANCE_COLORS,
} from "@/lib/types";
import {
  createMemory,
  updateMemory,
  deleteMemory,
  searchMemories,
} from "@/hooks/use-memories";

interface Props {
  memory: Memory | null;
  agents: Agent[];
  isNew?: boolean;
  onSaved: (m: Memory) => void;
  onDeleted: () => void;
}

export function MemoryEditor({ memory, agents, isNew, onSaved, onDeleted }: Props) {
  const [content, setContent] = useState("");
  const [category, setCategory] = useState("general");
  const [tagsInput, setTagsInput] = useState("");
  const [importance, setImportance] = useState<MemoryImportance>(2);
  const [agentId, setAgentId] = useState("");
  const [saving, setSaving] = useState(false);

  // 语义搜索
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<Memory[] | null>(null);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    if (isNew) {
      setContent("");
      setCategory("general");
      setTagsInput("");
      setImportance(2);
      setAgentId(agents.find((a) => !a.isHuman)?.id ?? "");
    } else if (memory) {
      setContent(memory.content);
      setCategory(memory.category);
      setTagsInput(memory.tags.join(", "));
      setImportance(memory.importance);
      setAgentId(memory.agentId);
    }
  }, [memory, isNew, agents]);

  const handleSave = async () => {
    if (!content.trim()) return;
    setSaving(true);
    const tags = tagsInput.split(",").map((t) => t.trim()).filter(Boolean);
    try {
      let result: Memory;
      if (isNew || !memory) {
        if (!agentId) {
          toast.error("请选择 Agent");
          setSaving(false);
          return;
        }
        result = await createMemory({ agentId, content, category, tags, importance });
      } else {
        result = await updateMemory(memory.id, { content, category, tags, importance });
      }
      toast.success("记忆已保存");
      onSaved(result);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!memory) return;
    toast("确认删除这条记忆？", {
      action: {
        label: "删除",
        onClick: async () => {
          try {
            await deleteMemory(memory.id);
            toast.success("记忆已删除");
            onDeleted();
          } catch {
            toast.error("删除失败");
          }
        },
      },
      cancel: { label: "取消", onClick: () => {} },
    });
  };

  const handleSearch = async () => {
    if (!searchQuery.trim()) return;
    setSearching(true);
    try {
      const res = await searchMemories(searchQuery, agentId || undefined, 10);
      setSearchResults(res.data);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "搜索失败");
    } finally {
      setSearching(false);
    }
  };

  if (!memory && !isNew) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center text-zinc-500 text-sm gap-4">
        <p>请从左侧选择一条记忆，或创建新记忆</p>
        <div className="w-80 space-y-2">
          <p className="text-zinc-400 text-xs">语义搜索</p>
          <div className="flex gap-2">
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSearch()}
              placeholder="输入自然语言查询..."
              className="flex-1 px-3 py-1.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
            <button
              onClick={handleSearch}
              disabled={searching}
              className="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm transition-colors"
            >
              <Search className="w-3.5 h-3.5" />
            </button>
          </div>
          {searchResults && (
            <div className="mt-2 space-y-1 max-h-60 overflow-y-auto">
              {searchResults.length === 0 ? (
                <p className="text-zinc-600 text-xs text-center py-2">无结果</p>
              ) : (
                searchResults.map((m) => (
                  <div key={m.id} className="p-2 bg-zinc-800/50 rounded text-xs text-zinc-300 line-clamp-2">
                    <span className={cn("inline-block px-1 py-0.5 rounded mr-1 text-xs", IMPORTANCE_COLORS[m.importance])}>
                      {IMPORTANCE_LABELS[m.importance]}
                    </span>
                    {m.content}
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      </div>
    );
  }

  const agentOptions = agents.filter((a) => !a.isHuman);

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-zinc-800 flex-shrink-0">
        <div className="flex-1" />
        <button
          onClick={handleSave}
          disabled={saving || !content.trim()}
          className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-xs font-medium transition-colors"
        >
          <Save className="w-3 h-3" />
          {saving ? "保存中..." : "保存"}
        </button>
        {!isNew && memory && (
          <button
            onClick={handleDelete}
            className="flex items-center gap-1 px-3 py-1.5 bg-red-600/20 hover:bg-red-600/40 text-red-400 rounded-md text-xs font-medium transition-colors"
          >
            <Trash2 className="w-3 h-3" />
            删除
          </button>
        )}
      </div>

      {/* Form */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Agent 选择（仅新建时） */}
        {isNew && (
          <div>
            <label className="block text-xs text-zinc-400 mb-1">Agent</label>
            <select
              value={agentId}
              onChange={(e) => setAgentId(e.target.value)}
              className="w-full py-1.5 px-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-200 text-sm focus:outline-none focus:border-blue-500"
            >
              <option value="">选择 Agent...</option>
              {agentOptions.map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
          </div>
        )}

        {/* 内容 */}
        <div>
          <label className="block text-xs text-zinc-400 mb-1">内容</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="记忆内容..."
            rows={6}
            className="w-full p-3 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-200 text-sm placeholder-zinc-600 focus:outline-none focus:border-blue-500 resize-none"
          />
        </div>

        {/* 分类 + 重要性 */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-xs text-zinc-400 mb-1">分类</label>
            <input
              value={category}
              onChange={(e) => setCategory(e.target.value)}
              placeholder="general"
              className="w-full py-1.5 px-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-200 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div>
            <label className="block text-xs text-zinc-400 mb-1">重要性</label>
            <div className="flex gap-1">
              {([0, 1, 2, 3, 4] as MemoryImportance[]).map((imp) => (
                <button
                  key={imp}
                  onClick={() => setImportance(imp)}
                  className={cn(
                    "flex-1 py-1.5 rounded text-xs transition-colors",
                    importance === imp
                      ? IMPORTANCE_COLORS[imp]
                      : "bg-zinc-800 text-zinc-500 hover:text-zinc-300"
                  )}
                >
                  {IMPORTANCE_LABELS[imp]}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* 标签 */}
        <div>
          <label className="block text-xs text-zinc-400 mb-1">标签（逗号分隔）</label>
          <input
            value={tagsInput}
            onChange={(e) => setTagsInput(e.target.value)}
            placeholder="tag1, tag2"
            className="w-full py-1.5 px-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-200 text-sm focus:outline-none focus:border-blue-500"
          />
        </div>

        {/* 元数据（仅查看已有记忆时显示） */}
        {!isNew && memory && (
          <div className="pt-4 border-t border-zinc-800 space-y-2 text-xs text-zinc-500">
            <div className="grid grid-cols-2 gap-2">
              <div>
                <span className="text-zinc-600">ID:</span>{" "}
                <span className="font-mono">{memory.id.slice(0, 8)}...</span>
              </div>
              <div>
                <span className="text-zinc-600">来源:</span> {memory.source}
              </div>
              <div>
                <span className="text-zinc-600">访问次数:</span> {memory.accessCount}
              </div>
              <div>
                <span className="text-zinc-600">Agent:</span>{" "}
                {agents.find((a) => a.id === memory.agentId)?.name ?? memory.agentId.slice(0, 8)}
              </div>
              <div>
                <span className="text-zinc-600">创建:</span>{" "}
                {new Date(memory.createdAt).toLocaleString("zh-CN")}
              </div>
              <div>
                <span className="text-zinc-600">更新:</span>{" "}
                {new Date(memory.updatedAt).toLocaleString("zh-CN")}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
