"use client";

import { useState, useEffect, useMemo } from "react";
import { toast } from "sonner";
import { Shell } from "@/components/layout/shell";
import { AgentList } from "@/components/agents/agent-list";
import { CreateAgentDialog } from "@/components/agents/create-agent-dialog";
import { useAgents } from "@/hooks/use-agents";
import { useModels } from "@/hooks/use-models";
import type { Agent } from "@/lib/types";

export default function AgentsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [isChairman, setIsChairman] = useState(false);
  const { agents, mutate } = useAgents();

  // 快速创建 HR
  const [isBootstrapping, setIsBootstrapping] = useState(false);
  const [bootstrapResult, setBootstrapResult] = useState<{ apiKey: string; agentId: string } | null>(null);
  const [showModelPicker, setShowModelPicker] = useState(false);
  const [selectedModel, setSelectedModel] = useState("");

  // 获取 nanoclaw 兼容模型（bootstrap HR 固定用 nanoclaw）
  const { models: nanoclawModels, isLoading: modelsLoading } = useModels("nanoclaw");

  const hasHR = useMemo(() =>
    agents.some(a => a.position === "hr_director" || a.position === "hr_manager"),
    [agents]
  );

  useEffect(() => {
    const token = localStorage.getItem("lc_token");
    if (!token) return;
    fetch("/api/v1/agents", { headers: { Authorization: `Bearer ${token}` } })
      .then(r => r.json())
      .then(d => {
        const agentId = localStorage.getItem("lc_agent_id");
        const me = (d.data as Agent[])?.find(a => a.id === agentId);
        if (me?.roleType === "chairman") setIsChairman(true);
      })
      .catch(() => {});
  }, []);

  function handleCreated(_agent: Agent, _apiKey: string) {
    mutate();
  }

  function handleStartBootstrap() {
    setSelectedModel("");
    setShowModelPicker(true);
  }

  async function handleBootstrapHR() {
    if (!selectedModel) {
      toast.error("请先选择模型");
      return;
    }
    setShowModelPicker(false);
    setIsBootstrapping(true);
    setBootstrapResult(null);
    try {
      const token = localStorage.getItem("lc_token");
      const headers = { "Content-Type": "application/json", Authorization: `Bearer ${token}` };

      // 1. 创建 HR Agent（含模型）
      const createRes = await fetch("/api/v1/agents", {
        method: "POST", headers,
        body: JSON.stringify({ position: "hr_director", model: selectedModel }),
      });
      const createData = await createRes.json();
      if (!createRes.ok) throw new Error(createData.error || "创建失败");

      const agentId = createData.agent.id as string;
      const apiKey = (createData.api_key || "") as string;

      // 2. 部署为本地 Docker (nanoclaw)
      const deployRes = await fetch(`/api/v1/agents/${agentId}/deploy`, {
        method: "POST", headers,
        body: JSON.stringify({ deployType: "local_docker", agentImage: "nanoclaw", apiKey }),
      });
      const deployData = await deployRes.json();
      if (!deployRes.ok) throw new Error(deployData.error || "部署失败");

      setBootstrapResult({ apiKey, agentId });
      toast.success("HR Agent 已创建并部署");
      mutate();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "创建失败");
    } finally {
      setIsBootstrapping(false);
    }
  }

  return (
    <Shell>
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-semibold text-zinc-50">Agent 管理</h1>
            <p className="text-zinc-400 text-sm mt-1">管理您的 AI Agent 及部署状态</p>
          </div>
          {isChairman && (
            <button
              onClick={() => setCreateOpen(true)}
              className="flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              创建 Agent
            </button>
          )}
        </div>

        {/* 无 HR 时显示引导卡片 */}
        {isChairman && !hasHR && !bootstrapResult && (
          <div className="border border-amber-500/30 bg-amber-500/5 rounded-xl p-6">
            <div className="flex items-start gap-4">
              <div className="text-3xl">👤</div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-zinc-100">创建第一个 HR Agent</h3>
                <p className="text-sm text-zinc-400 mt-1">
                  HR Agent 是公司的人力资源管理者，负责招聘、部署和管理其他 Agent。
                  创建后将自动以 NanoClaw 镜像在本地 Docker 启动。
                </p>
                <button
                  onClick={handleStartBootstrap}
                  disabled={isBootstrapping}
                  className="mt-4 px-5 py-2.5 rounded-lg bg-amber-600 hover:bg-amber-700 text-sm font-medium text-white transition-colors disabled:opacity-50"
                >
                  {isBootstrapping ? "创建并部署中…" : "一键创建 HR Agent"}
                </button>
              </div>
            </div>
          </div>
        )}

        {/* 创建成功提示 */}
        {bootstrapResult && (
          <div className="border border-green-500/30 bg-green-500/5 rounded-xl p-6 space-y-3">
            <div className="flex items-center gap-2 text-green-400">
              <span className="text-lg">✓</span>
              <span className="font-semibold">HR Agent 已创建并部署</span>
            </div>
            <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4">
              <p className="text-xs text-yellow-400 font-medium mb-2">API Key（仅显示一次，请妥善保存）</p>
              <code className="text-xs font-mono text-yellow-300 break-all">{bootstrapResult.apiKey}</code>
            </div>
            <p className="text-sm text-zinc-400">
              HR Agent 正在启动，稍后将自动连接到系统并取名。后续 Agent 的创建和部署可委托给 HR 完成。
            </p>
            <button
              onClick={() => setBootstrapResult(null)}
              className="px-4 py-2 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-sm text-zinc-200 transition-colors"
            >
              知道了
            </button>
          </div>
        )}

        <AgentList onOpenCreate={() => setCreateOpen(true)} isChairman={isChairman} />
      </div>

      <CreateAgentDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={handleCreated}
        hasHR={hasHR}
      />

      {/* 模型选择弹窗 */}
      {showModelPicker && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-sm mx-4 shadow-2xl">
            <div className="flex items-center justify-between px-5 pt-5 pb-3 border-b border-zinc-800">
              <h3 className="text-base font-semibold text-zinc-50">选择 LLM 模型</h3>
              <button onClick={() => setShowModelPicker(false)} className="text-zinc-500 hover:text-zinc-300 transition-colors text-xl leading-none">&times;</button>
            </div>
            <div className="p-5 space-y-3">
              <p className="text-xs text-zinc-500">
                HR Agent 使用 NanoClaw 框架，仅支持 Anthropic API 格式的模型。
              </p>
              {modelsLoading ? (
                <div className="text-xs text-zinc-500 py-2">加载模型列表...</div>
              ) : nanoclawModels.length === 0 ? (
                <div className="text-xs text-amber-400 py-2">
                  没有可用的 Anthropic 格式模型，请先在 LLM Gateway 中配置 Provider。
                </div>
              ) : (
                <div className="space-y-2">
                  {nanoclawModels.map(m => (
                    <button
                      key={m}
                      onClick={() => setSelectedModel(m)}
                      className={`w-full px-4 py-3 rounded-lg border text-left text-sm transition-colors ${
                        selectedModel === m
                          ? "border-amber-500 bg-amber-500/10 text-zinc-100"
                          : "border-zinc-700 bg-zinc-800 hover:border-zinc-600 text-zinc-300"
                      }`}
                    >
                      <div className="flex items-center gap-3">
                        <div className={`w-3.5 h-3.5 rounded-full border-2 flex-shrink-0 ${selectedModel === m ? "border-amber-500 bg-amber-500" : "border-zinc-600"}`} />
                        <span className="font-mono text-sm">{m}</span>
                      </div>
                    </button>
                  ))}
                </div>
              )}
            </div>
            <div className="flex justify-end gap-3 px-5 pb-5">
              <button onClick={() => setShowModelPicker(false)}
                className="px-4 py-2 rounded-lg text-sm text-zinc-400 hover:text-zinc-200 transition-colors">取消</button>
              <button onClick={handleBootstrapHR}
                disabled={!selectedModel || isBootstrapping}
                className="px-5 py-2 rounded-lg bg-amber-600 hover:bg-amber-700 text-sm font-medium text-white transition-colors disabled:opacity-40">
                {isBootstrapping ? "创建中…" : "确认创建"}
              </button>
            </div>
          </div>
        </div>
      )}
    </Shell>
  );
}
