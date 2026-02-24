"use client";

import { useState } from "react";
import { Shell } from "@/components/layout/shell";
import { PromptNav, type PromptSelection } from "@/components/prompts/prompt-nav";
import { PromptEditor } from "@/components/prompts/prompt-editor";
import { usePrompts } from "@/hooks/use-prompts";

export default function PromptsPage() {
  const { data, isLoading, mutate } = usePrompts();
  const [selected, setSelected] = useState<PromptSelection | null>(null);

  return (
    <Shell noPadding>
      <div className="flex h-[calc(100vh-3.5rem)]">
        <div className="w-72 flex-shrink-0">
          <PromptNav data={data} selected={selected} onSelect={setSelected} />
        </div>
        <PromptEditor
          selected={selected}
          data={data}
          onSaved={() => mutate()}
        />
      </div>
    </Shell>
  );
}
