"use client";

import { useState } from "react";
import { Shell } from "@/components/layout/shell";
import { KnowledgeList } from "@/components/knowledge/knowledge-list";
import { KnowledgeEditor } from "@/components/knowledge/knowledge-editor";
import { useKnowledgeDocs, KnowledgeDoc } from "@/hooks/use-knowledge";

export default function KnowledgePage() {
  const { docs, isLoading, refresh } = useKnowledgeDocs();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [isNew, setIsNew] = useState(false);

  const selectedDoc = docs.find((d) => d.id === selectedId) ?? null;

  const handleNew = () => {
    setSelectedId(null);
    setIsNew(true);
  };

  const handleSaved = (doc: KnowledgeDoc) => {
    refresh();
    setSelectedId(doc.id);
    setIsNew(false);
  };

  const handleDeleted = () => {
    refresh();
    setSelectedId(null);
    setIsNew(false);
  };

  return (
    <Shell noPadding>
      <div className="flex h-[calc(100vh-3.5rem)]">
        <div className="w-64 flex-shrink-0">
          <KnowledgeList
            docs={docs}
            selectedId={selectedId}
            onSelect={(id) => { setSelectedId(id); setIsNew(false); }}
            onNew={handleNew}
            isLoading={isLoading}
          />
        </div>
        <KnowledgeEditor
          doc={isNew ? null : selectedDoc}
          isNew={isNew}
          onSaved={handleSaved}
          onDeleted={handleDeleted}
        />
      </div>
    </Shell>
  );
}
