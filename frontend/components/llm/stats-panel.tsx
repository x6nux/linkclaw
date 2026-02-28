"use client";

import { LLMUsageStats, LLMDailyUsage, LLMRecentLog } from "@/lib/types";
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip,
  ResponsiveContainer, Legend,
} from "recharts";
function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const mins = Math.floor(diff / 60_000);
  if (mins < 1) return "刚刚";
  if (mins < 60) return `${mins} 分钟前`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} 小时前`;
  return `${Math.floor(hours / 24)} 天前`;
}

// ===== 数字格式化 =====

function fmtTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function fmtUSD(usd: number): string {
  return `$${usd.toFixed(4)}`;
}

// ===== Provider 统计卡片列表 =====

export function ProviderStats({ stats }: { stats: LLMUsageStats[] }) {
  if (!stats?.length) return null;

  return (
    <div>
      <h2 className="text-zinc-50 font-semibold mb-4">用量统计（按 Provider）</h2>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-zinc-800 text-zinc-400 text-xs">
              <th className="text-left py-2 pr-4">Provider</th>
              <th className="text-right py-2 pr-4">请求数</th>
              <th className="text-right py-2 pr-4">成功率</th>
              <th className="text-right py-2 pr-4">Input Tokens</th>
              <th className="text-right py-2 pr-4">Output Tokens</th>
              <th className="text-right py-2 pr-4">Cache 创建</th>
              <th className="text-right py-2 pr-4">Cache 命中</th>
              <th className="text-right py-2">总费用</th>
            </tr>
          </thead>
          <tbody>
            {stats.map((s) => (
              <tr key={s.provider_id} className="border-b border-zinc-800/50 hover:bg-zinc-900/50">
                <td className="py-2.5 pr-4 text-zinc-200 font-medium">{s.provider_name}</td>
                <td className="py-2.5 pr-4 text-right text-zinc-300">{s.total_requests.toLocaleString()}</td>
                <td className="py-2.5 pr-4 text-right">
                  <span className={
                    s.total_requests > 0 && s.success_requests / s.total_requests >= 0.95
                      ? "text-green-400"
                      : "text-amber-400"
                  }>
                    {s.total_requests > 0
                      ? `${((s.success_requests / s.total_requests) * 100).toFixed(1)}%`
                      : "—"}
                  </span>
                </td>
                <td className="py-2.5 pr-4 text-right text-zinc-300 font-mono">{fmtTokens(s.input_tokens)}</td>
                <td className="py-2.5 pr-4 text-right text-zinc-300 font-mono">{fmtTokens(s.output_tokens)}</td>
                <td className="py-2.5 pr-4 text-right text-zinc-400 font-mono">{fmtTokens(s.cache_creation_tokens)}</td>
                <td className="py-2.5 pr-4 text-right text-green-400 font-mono">{fmtTokens(s.cache_read_tokens)}</td>
                <td className="py-2.5 text-right text-blue-400 font-mono">{fmtUSD(s.total_cost_usd)}</td>
              </tr>
            ))}
          </tbody>
          <tfoot>
            <tr className="text-zinc-400 text-xs font-medium">
              <td className="pt-3 pr-4">合计</td>
              <td className="pt-3 pr-4 text-right">{stats.reduce((s, r) => s + r.total_requests, 0).toLocaleString()}</td>
              <td className="pt-3 pr-4" />
              <td className="pt-3 pr-4 text-right font-mono">{fmtTokens(stats.reduce((s, r) => s + r.input_tokens, 0))}</td>
              <td className="pt-3 pr-4 text-right font-mono">{fmtTokens(stats.reduce((s, r) => s + r.output_tokens, 0))}</td>
              <td className="pt-3 pr-4 text-right font-mono">{fmtTokens(stats.reduce((s, r) => s + r.cache_creation_tokens, 0))}</td>
              <td className="pt-3 pr-4 text-right font-mono text-green-400">{fmtTokens(stats.reduce((s, r) => s + r.cache_read_tokens, 0))}</td>
              <td className="pt-3 text-right font-mono text-blue-400">{fmtUSD(stats.reduce((s, r) => s + r.total_cost_usd, 0))}</td>
            </tr>
          </tfoot>
        </table>
      </div>
    </div>
  );
}

// ===== 每日用量折线图 =====

export function DailyUsageChart({ daily }: { daily: LLMDailyUsage[] }) {
  if (!daily?.length) return null;

  // 取最近 30 天，格式化日期标签
  const chartData = daily.map((d) => ({
    date: d.date.slice(5), // MM-DD
    输入: Math.round(d.input_tokens / 1000),
    输出: Math.round(d.output_tokens / 1000),
    费用: Number(d.cost_usd.toFixed(4)),
    请求: d.requests,
  }));

  return (
    <div>
      <h2 className="text-zinc-50 font-semibold mb-4">每日用量（近 30 天）</h2>
      <div className="h-56">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
            <XAxis dataKey="date" tick={{ fontSize: 11, fill: "#71717a" }} />
            <YAxis tick={{ fontSize: 11, fill: "#71717a" }} unit="K" />
            <Tooltip
              contentStyle={{ background: "#18181b", border: "1px solid #3f3f46", borderRadius: 6, fontSize: 12 }}
              labelStyle={{ color: "#a1a1aa" }}
            />
            <Legend wrapperStyle={{ fontSize: 12, color: "#a1a1aa" }} />
            <Line type="monotone" dataKey="输入" stroke="#3b82f6" dot={false} strokeWidth={1.5} />
            <Line type="monotone" dataKey="输出" stroke="#10b981" dot={false} strokeWidth={1.5} />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ===== 最近请求日志 =====

export function RecentLogs({ logs }: { logs: LLMRecentLog[] }) {
  if (!logs?.length) return null;

  return (
    <div>
      <h2 className="text-zinc-50 font-semibold mb-4">最近请求</h2>
      <div className="space-y-1">
        {logs.map((log) => (
          <div
            key={log.id}
            className="flex items-center gap-3 px-3 py-2 rounded bg-zinc-900 border border-zinc-800/50 text-xs"
          >
            <span className={
              log.status === "success" ? "w-1.5 h-1.5 rounded-full bg-green-500 flex-shrink-0" :
              "w-1.5 h-1.5 rounded-full bg-red-500 flex-shrink-0"
            } />
            <span className="font-mono text-zinc-400 w-36 truncate">{log.request_model}</span>
            <span className="text-zinc-300 font-mono">
              {fmtTokens(log.input_tokens)}↑ {fmtTokens(log.output_tokens)}↓
            </span>
            {log.cache_read_tokens > 0 && (
              <span className="text-green-400 font-mono">{fmtTokens(log.cache_read_tokens)} cached</span>
            )}
            <span className="text-zinc-500 font-mono">{log.latency_ms ? `${log.latency_ms}ms` : "—"}</span>
            {log.retry_count > 0 && (
              <span className="text-amber-400">重试 {log.retry_count}</span>
            )}
            {log.error_msg && (
              <span className="text-red-400 truncate max-w-32">{log.error_msg}</span>
            )}
            <span className="ml-auto text-zinc-600">
              {relativeTime(log.created_at)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
