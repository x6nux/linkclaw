"use client";

import { Shell } from "@/components/layout/shell";
import { IndexTasks } from "@/components/context/index-tasks";
import { SearchForm } from "@/components/context/search-form";

export default function ContextPage() {
  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">上下文索引</h1>
          <p className="text-zinc-400 text-sm mt-1">代码仓库索引与语义搜索</p>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <IndexTasks />
          <SearchForm />
        </div>
      </div>
    </Shell>
  );
}
