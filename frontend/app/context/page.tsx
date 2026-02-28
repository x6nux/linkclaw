"use client";

import { useState } from "react";
import { Shell } from "@/components/layout/shell";
import { IndexTasks } from "@/components/context/index-tasks";
import { SearchModal } from "@/components/context/search-modal";

export default function ContextPage() {
  const [searchOpen, setSearchOpen] = useState(false);

  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">上下文索引</h1>
          <p className="text-zinc-400 text-sm mt-1">代码仓库索引与语义搜索</p>
        </div>
        <IndexTasks onOpenSearch={() => setSearchOpen(true)} />
      </div>
      <SearchModal open={searchOpen} onClose={() => setSearchOpen(false)} />
    </Shell>
  );
}
