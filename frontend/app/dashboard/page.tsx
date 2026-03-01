"use client";

import Link from "next/link";
import { Shell } from "@/components/layout/shell";
import { useAgents } from "@/hooks/use-agents";
import { useTasks } from "@/hooks/use-tasks";
import { useKnowledgeDocs } from "@/hooks/use-knowledge";
import { Bot } from "lucide-react";

function StatSkeleton() {
  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 animate-pulse">
      <div className="h-4 w-20 bg-zinc-800 rounded" />
      <div className="h-8 w-12 bg-zinc-800 rounded mt-2" />
      <div className="h-3 w-16 bg-zinc-800/60 rounded mt-2" />
    </div>
  );
}

export default function DashboardPage() {
  const { agents, isLoading: agentsLoading } = useAgents();
  const { tasks, isLoading: tasksLoading } = useTasks();
  const { docs, isLoading: docsLoading } = useKnowledgeDocs();

  const isLoading = agentsLoading || tasksLoading || docsLoading;

  const onlineCount = agents?.filter((a) => a.status === "online").length ?? 0;
  const activeTaskCount =
    tasks?.filter(
      (t) => t.status === "in_progress" || t.status === "assigned"
    ).length ?? 0;

  const stats = [
    { label: "在线 Agent", value: String(onlineCount), desc: "当前活跃" },
    { label: "进行中任务", value: String(activeTaskCount), desc: "需要处理" },
    { label: "今日消息", value: "—", desc: "统计开发中" },
    {
      label: "知识文档",
      value: String(docs.length),
      desc: "总计",
    },
  ];

  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">概览</h1>
          <p className="text-zinc-400 text-sm mt-1">监控您的 AI Agent 虚拟公司</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          {isLoading
            ? Array.from({ length: 4 }).map((_, i) => <StatSkeleton key={i} />)
            : stats.map(({ label, value, desc }, idx) => (
                <div
                  key={label}
                  className="group relative bg-zinc-900/80 backdrop-blur-sm border border-zinc-800/50 hover:border-zinc-700/50 transition-all duration-300 rounded-lg p-4 hover-lift overflow-hidden"
                >
                  {/* Gradient accent on hover */}
                  <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-indigo-500/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />

                  {/* Content */}
                  <div className="relative">
                    <div className="flex items-center gap-2">
                      <div className={cn(
                        "w-1 h-6 rounded-full transition-all duration-300",
                        idx === 0 ? "bg-gradient-to-b from-emerald-500 to-emerald-600" :
                        idx === 1 ? "bg-gradient-to-b from-blue-500 to-blue-600" :
                        idx === 2 ? "bg-gradient-to-b from-amber-500 to-amber-600" :
                        "bg-gradient-to-b from-purple-500 to-purple-600"
                      )} />
                      <div className="text-zinc-400 text-sm">{label}</div>
                    </div>
                    <div className="text-3xl font-semibold text-zinc-50 mt-2 tracking-tight">
                      {value}
                    </div>
                    <div className="text-zinc-500 text-xs mt-2 flex items-center gap-1">
                      <span className="inline-block w-1.5 h-1.5 rounded-full bg-zinc-600" />
                      {desc}
                    </div>
                  </div>
                </div>
              ))}
        </div>

        {!isLoading && onlineCount === 0 && activeTaskCount === 0 && (
          <div className="relative bg-zinc-900/50 backdrop-blur-sm border border-zinc-800/50 rounded-lg p-8 flex flex-col items-center gap-3 overflow-hidden">
            {/* Background gradient */}
            <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 via-transparent to-purple-500/5" />

            <div className="relative">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-zinc-800 to-zinc-900 border border-zinc-700/50 flex items-center justify-center shadow-lg shadow-black/20">
                <Bot className="w-8 h-8 text-zinc-500" />
              </div>
              <div className="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-zinc-600 animate-pulse" />
            </div>

            <p className="relative text-zinc-400 text-sm text-center max-w-md">
              暂无活动数据。请先创建一个 Agent 并通过 MCP 连接。
            </p>
            <Link
              href="/agents"
              className="relative inline-flex items-center gap-1.5 px-4 py-2 rounded-md text-sm font-medium text-blue-400 bg-blue-500/10 border border-blue-500/20 hover:bg-blue-500/15 hover:border-blue-500/40 hover:text-blue-300 transition-all duration-200 hover-lift"
            >
              前往创建 Agent
              <span className="text-lg leading-none">&rarr;</span>
            </Link>
          </div>
        )}
      </div>
    </Shell>
  );
}
