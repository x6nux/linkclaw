"use client";

import { useState } from "react";
import { Search, Plus, Brain, X, Trash2, Filter } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  Memory,
  MemoryImportance,
  IMPORTANCE_LABELS,
  IMPORTANCE_COLORS,
} from "@/lib/types";
import { Agent } from "@/lib/types";

interface Props {
  memories: Memory[];
  agents: Agent[];
  selectedId: string | null;
  selectedIds: Set<string>;
  onSelect: (id: string) => void;
  onToggleSelect: (id: string) => void;
  onNew: () => void;
  onBatchDelete: () => void;
  onAgentFilter: (agentId: string) => void;
  onImportanceFilter: (importance: number | undefined) => void;
  agentFilter: string;
  importanceFilter: number | undefined;
  isLoading?: boolean;
}

function ListSkeleton() {
  return (
    <div className="divide-y divide-zinc-800/50">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="px-3 py-3 animate-pulse">
          <div className="space-y-2">
            <div className="h-4 w-3/4 bg-zinc-800 rounded" />
            <div className="h-3 w-1/2 bg-zinc-800/60 rounded" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function MemoryList({
  memories,
  agents,
  selectedId,
  selectedIds,
  onSelect,
  onToggleSelect,
  onNew,
  onBatchDelete,
  onAgentFilter,
  onImportanceFilter,
  agentFilter,
  importanceFilter,
  isLoading,
}: Props) {
  const [search, setSearch] = useState("");
  const [showFilters, setShowFilters] = useState(false);

  const filtered = memories.filter((m) =>
    m.content.toLowerCase().includes(search.toLowerCase()) ||
    m.category.toLowerCase().includes(search.toLowerCase()) ||
    m.tags.some((t) => t.toLowerCase().includes(search.toLowerCase()))
  );

  const agentName = (id: string) =>
    agents.find((a) => a.id === id)?.name ?? id.slice(0, 8);

  return (
    <div className="flex flex-col h-full border-r border-zinc-800 w-80 flex-shrink-0">
      <div className="p-3 border-b border-zinc-800 space-y-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 w-3.5 h-3.5 text-zinc-500" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索记忆..."
            className={cn(
              "w-full pl-8 py-1.5 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500",
              search ? "pr-8" : "pr-3"
            )}
          />
          {search && (
            <button
              onClick={() => setSearch("")}
              className="absolute right-2 top-2 text-zinc-500 hover:text-zinc-300 transition-colors"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>

        <div className="flex gap-2">
          <button
            onClick={onNew}
            className="flex-1 flex items-center justify-center gap-1.5 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-sm font-medium transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            新建
          </button>
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={cn(
              "px-2.5 py-1.5 rounded-md text-sm transition-colors",
              showFilters || agentFilter || importanceFilter !== undefined
                ? "bg-blue-500/20 text-blue-400"
                : "bg-zinc-800 text-zinc-400 hover:text-zinc-200"
            )}
          >
            <Filter className="w-3.5 h-3.5" />
          </button>
        </div>

        {showFilters && (
          <div className="space-y-2 pt-1">
            <select
              value={agentFilter}
              onChange={(e) => onAgentFilter(e.target.value)}
              className="w-full py-1.5 px-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-200 text-sm focus:outline-none focus:border-blue-500"
            >
              <option value="">全部 Agent</option>
              {agents.filter((a) => !a.is_human).map((a) => (
                <option key={a.id} value={a.id}>{a.name}</option>
              ))}
            </select>
            <div className="flex flex-wrap gap-1">
              {([0, 1, 2, 3, 4] as MemoryImportance[]).map((imp) => (
                <button
                  key={imp}
                  onClick={() => onImportanceFilter(importanceFilter === imp ? undefined : imp)}
                  className={cn(
                    "px-2 py-0.5 rounded text-xs transition-colors",
                    importanceFilter === imp
                      ? IMPORTANCE_COLORS[imp]
                      : "bg-zinc-800 text-zinc-500 hover:text-zinc-300"
                  )}
                >
                  {IMPORTANCE_LABELS[imp]}
                </button>
              ))}
            </div>
          </div>
        )}

        {selectedIds.size > 0 && (
          <button
            onClick={onBatchDelete}
            className="w-full flex items-center justify-center gap-1.5 py-1.5 bg-red-600/20 hover:bg-red-600/40 text-red-400 rounded-md text-sm transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
            删除选中 ({selectedIds.size})
          </button>
        )}
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <ListSkeleton />
        ) : filtered.length === 0 ? (
          <div className="p-6 flex flex-col items-center gap-2 text-center">
            <Brain className="w-6 h-6 text-zinc-600" />
            <p className="text-zinc-500 text-sm">
              {search ? `未找到匹配的记忆` : "暂无记忆"}
            </p>
          </div>
        ) : (
          filtered.map((mem) => (
            <div
              key={mem.id}
              className={cn(
                "w-full text-left px-3 py-2.5 border-b border-zinc-800/50 hover:bg-zinc-800/50 transition-colors cursor-pointer",
                selectedId === mem.id && "bg-blue-500/10 border-l-2 border-l-blue-500"
              )}
            >
              <div className="flex items-start gap-2">
                <input
                  type="checkbox"
                  checked={selectedIds.has(mem.id)}
                  onChange={() => onToggleSelect(mem.id)}
                  className="mt-1 flex-shrink-0 accent-blue-500"
                  onClick={(e) => e.stopPropagation()}
                />
                <div className="min-w-0 flex-1" onClick={() => onSelect(mem.id)}>
                  <div className="text-sm text-zinc-200 line-clamp-2">{mem.content}</div>
                  <div className="flex items-center gap-1.5 mt-1 flex-wrap">
                    <span className={cn("text-xs px-1.5 py-0.5 rounded", IMPORTANCE_COLORS[mem.importance])}>
                      {IMPORTANCE_LABELS[mem.importance]}
                    </span>
                    <span className="text-xs bg-zinc-700/50 text-zinc-400 px-1.5 py-0.5 rounded">
                      {mem.category}
                    </span>
                    <span className="text-xs text-zinc-600 ml-auto">
                      {agentName(mem.agent_id)}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
