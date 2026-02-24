"use client";

import { useState } from "react";
import { Search, Plus, FileText, X, BookOpen } from "lucide-react";
import { cn } from "@/lib/utils";
import { KnowledgeDoc } from "@/hooks/use-knowledge";

interface Props {
  docs: KnowledgeDoc[];
  selectedId: string | null;
  onSelect: (id: string) => void;
  onNew: () => void;
  isLoading?: boolean;
}

function ListSkeleton() {
  return (
    <div className="divide-y divide-zinc-800/50">
      {Array.from({ length: 4 }).map((_, i) => (
        <div key={i} className="px-3 py-3 animate-pulse">
          <div className="flex items-start gap-2">
            <div className="w-3.5 h-3.5 bg-zinc-800 rounded mt-0.5 flex-shrink-0" />
            <div className="flex-1 min-w-0 space-y-2">
              <div className="h-4 w-3/4 bg-zinc-800 rounded" />
              <div className="flex gap-1">
                <div className="h-4 w-10 bg-zinc-800/60 rounded" />
                <div className="h-4 w-12 bg-zinc-800/60 rounded" />
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function KnowledgeList({
  docs,
  selectedId,
  onSelect,
  onNew,
  isLoading,
}: Props) {
  const [search, setSearch] = useState("");

  const filtered = docs.filter(
    (d) =>
      d.title.toLowerCase().includes(search.toLowerCase()) ||
      d.tags.some((t) => t.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <div className="flex flex-col h-full border-r border-zinc-800">
      <div className="p-3 border-b border-zinc-800 space-y-2">
        <div className="relative">
          <Search className="absolute left-2.5 top-2.5 w-3.5 h-3.5 text-zinc-500" />
          <input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="搜索文档..."
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
        <button
          onClick={onNew}
          className="w-full flex items-center justify-center gap-2 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-md text-sm font-medium transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          新建文档
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <ListSkeleton />
        ) : filtered.length === 0 ? (
          <div className="p-6 flex flex-col items-center gap-2 text-center">
            <BookOpen className="w-6 h-6 text-zinc-600" />
            <p className="text-zinc-500 text-sm">
              {search ? `未找到与 "${search}" 匹配的文档` : "还没有文档"}
            </p>
            {!search && (
              <p className="text-zinc-600 text-xs">
                点击上方「新建文档」开始添加知识库内容
              </p>
            )}
          </div>
        ) : (
          filtered.map((doc) => (
            <button
              key={doc.id}
              onClick={() => onSelect(doc.id)}
              className={cn(
                "w-full text-left px-3 py-3 border-b border-zinc-800/50 hover:bg-zinc-800/50 transition-colors",
                selectedId === doc.id && "bg-blue-500/10 border-l-2 border-l-blue-500"
              )}
            >
              <div className="flex items-start gap-2">
                <FileText className="w-3.5 h-3.5 text-zinc-500 mt-0.5 flex-shrink-0" />
                <div className="min-w-0">
                  <div className="text-sm text-zinc-200 truncate">{doc.title}</div>
                  {doc.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {doc.tags.slice(0, 3).map((tag) => (
                        <span
                          key={tag}
                          className="text-xs bg-zinc-700 text-zinc-300 px-1.5 py-0.5 rounded"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>
            </button>
          ))
        )}
      </div>
    </div>
  );
}
