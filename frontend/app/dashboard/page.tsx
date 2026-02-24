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
            : stats.map(({ label, value, desc }) => (
                <div
                  key={label}
                  className="bg-zinc-900 border border-zinc-800 hover:border-zinc-700 transition-colors rounded-lg p-4"
                >
                  <div className="text-zinc-400 text-sm">{label}</div>
                  <div className="text-3xl font-semibold text-zinc-50 mt-1">
                    {value}
                  </div>
                  <div className="text-zinc-500 text-xs mt-1">{desc}</div>
                </div>
              ))}
        </div>

        {!isLoading && onlineCount === 0 && activeTaskCount === 0 && (
          <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-8 flex flex-col items-center gap-3">
            <Bot className="w-8 h-8 text-zinc-600" />
            <p className="text-zinc-400 text-sm text-center">
              暂无活动数据。请先创建一个 Agent 并通过 MCP 连接。
            </p>
            <Link
              href="/agents"
              className="text-sm text-blue-400 hover:text-blue-300 transition-colors"
            >
              前往创建 Agent &rarr;
            </Link>
          </div>
        )}
      </div>
    </Shell>
  );
}
