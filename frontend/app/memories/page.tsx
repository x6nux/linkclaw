"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Shell } from "@/components/layout/shell";
import { MemoryList } from "@/components/memories/memory-list";
import { MemoryEditor } from "@/components/memories/memory-editor";
import { useMemories, batchDeleteMemories } from "@/hooks/use-memories";
import { useAgents } from "@/hooks/use-agents";
import { Memory } from "@/lib/types";

export default function MemoriesPage() {
  const [agentFilter, setAgentFilter] = useState("");
  const [importanceFilter, setImportanceFilter] = useState<number | undefined>(undefined);
  const { memories, isLoading, refresh } = useMemories({
    agentId: agentFilter || undefined,
    importance: importanceFilter,
    limit: 100,
  });
  const { agents } = useAgents();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [isNew, setIsNew] = useState(false);

  const selectedMemory = memories.find((m) => m.id === selectedId) ?? null;

  const handleNew = () => {
    setSelectedId(null);
    setIsNew(true);
  };

  const handleSaved = (m: Memory) => {
    refresh();
    setSelectedId(m.id);
    setIsNew(false);
  };

  const handleDeleted = () => {
    refresh();
    setSelectedId(null);
    setIsNew(false);
  };

  const handleToggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleBatchDelete = () => {
    const ids = Array.from(selectedIds);
    toast(`确认删除 ${ids.length} 条记忆？`, {
      action: {
        label: "删除",
        onClick: async () => {
          try {
            await batchDeleteMemories(ids);
            toast.success(`已删除 ${ids.length} 条记忆`);
            setSelectedIds(new Set());
            if (selectedId && ids.includes(selectedId)) {
              setSelectedId(null);
            }
            refresh();
          } catch {
            toast.error("批量删除失败");
          }
        },
      },
      cancel: { label: "取消", onClick: () => {} },
    });
  };

  return (
    <Shell noPadding>
      <div className="flex h-[calc(100vh-3.5rem)]">
        <MemoryList
          memories={memories}
          agents={agents}
          selectedId={selectedId}
          selectedIds={selectedIds}
          onSelect={(id) => { setSelectedId(id); setIsNew(false); }}
          onToggleSelect={handleToggleSelect}
          onNew={handleNew}
          onBatchDelete={handleBatchDelete}
          onAgentFilter={setAgentFilter}
          onImportanceFilter={setImportanceFilter}
          agentFilter={agentFilter}
          importanceFilter={importanceFilter}
          isLoading={isLoading}
        />
        <MemoryEditor
          memory={isNew ? null : selectedMemory}
          agents={agents}
          isNew={isNew}
          onSaved={handleSaved}
          onDeleted={handleDeleted}
        />
      </div>
    </Shell>
  );
}
