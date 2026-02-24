"use client";

import type { Agent } from "@/lib/types";

type AgentImageType = "nanoclaw" | "openclaw";
type DeployMode = "manual" | "local_docker" | "ssh_docker";

const AGENT_IMAGE_INFO: Record<AgentImageType, { label: string }> = {
  nanoclaw: { label: "NanoClaw" },
  openclaw: { label: "OpenClaw" },
};

interface Props {
  createdAgent: Agent;
  apiKey: string;
  deployMode: DeployMode;
  agentImage: AgentImageType;
  deployResult: { status: string; error?: string } | null;
  isDeploying: boolean;
  onDeploy: () => void;
}

export function AgentStepResult({
  createdAgent, apiKey, deployMode, agentImage, deployResult, isDeploying, onDeploy,
}: Props) {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2 text-green-400">
        <span className="text-lg">{"\u2713"}</span>
        <span className="font-medium">Agent 创建成功</span>
      </div>

      {apiKey && (
        <div className="bg-yellow-500/10 border border-yellow-500/30 rounded-lg p-4">
          <p className="text-xs text-yellow-400 font-medium mb-2">API Key（仅显示一次，请妥善保存）</p>
          <code className="text-xs font-mono text-yellow-300 break-all">{apiKey}</code>
        </div>
      )}

      {deployMode === "manual" && (
        <div className="bg-zinc-800 rounded-lg p-4 text-sm text-zinc-400">
          <p>使用以上 API Key 配置 nanoclaw 或 openclaw，设置以下环境变量：</p>
          <pre className="mt-2 text-xs font-mono text-zinc-300 whitespace-pre-wrap">
{`AGENT_API_KEY=${apiKey}
LINKCLAW_MCP_URL=<your-mcp-url>/mcp/sse
ANTHROPIC_BASE_URL=<linkclaw-url>
ANTHROPIC_MODEL=glm-4.7`}
          </pre>
        </div>
      )}

      {(deployMode === "local_docker" || deployMode === "ssh_docker") && !deployResult && (
        <div>
          <p className="text-sm text-zinc-400 mb-3">
            {deployMode === "local_docker"
              ? "点击部署，后台将自动拉取镜像并启动容器"
              : `点击部署，将 SSH 到目标服务器启动容器`}
          </p>
          <button
            onClick={onDeploy}
            disabled={isDeploying}
            className="w-full py-2.5 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors disabled:opacity-50"
          >
            {isDeploying ? "部署中…" : `部署 ${AGENT_IMAGE_INFO[agentImage].label}`}
          </button>
        </div>
      )}

      {deployResult && (() => {
        const ok = deployResult.status !== "failed";
        return (
          <div className={`rounded-lg p-4 border ${ok ? "bg-green-500/10 border-green-500/30" : "bg-red-500/10 border-red-500/30"}`}>
            <p className={`text-sm font-medium ${ok ? "text-green-400" : "text-red-400"}`}>
              {ok ? "部署任务已提交，HR Agent 将自动完成后续启动" : "部署失败"}
            </p>
            {deployResult.error && <p className="text-xs text-red-400 mt-1">{deployResult.error}</p>}
          </div>
        );
      })()}
    </div>
  );
}
