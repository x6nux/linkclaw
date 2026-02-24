"use client";

import { useModels } from "@/hooks/use-models";

export type AgentImageType = "nanoclaw" | "openclaw";

const AGENT_IMAGE_INFO: Record<AgentImageType, { label: string; desc: string }> = {
  nanoclaw: { label: "NanoClaw", desc: "官方版本，内置更多工具，稳定性高" },
  openclaw: { label: "OpenClaw", desc: "完全开源，社区驱动，灵活可定制" },
};

interface Props {
  agentImage: AgentImageType;
  setAgentImage: (v: AgentImageType) => void;
  model: string;
  setModel: (v: string) => void;
}

export function AgentStepType({ agentImage, setAgentImage, model, setModel }: Props) {
  const { models, isLoading } = useModels(agentImage);

  return (
    <div className="space-y-4">
      <div className="space-y-3">
        <p className="text-sm text-zinc-400">选择 Agent 运行框架：</p>
        {(["nanoclaw", "openclaw"] as AgentImageType[]).map(img => (
          <button
            key={img}
            onClick={() => {
              setAgentImage(img);
              setModel("");  // 切换框架时重置模型选择
            }}
            className={`w-full p-4 rounded-lg border text-left transition-colors ${
              agentImage === img
                ? "border-blue-500 bg-blue-500/10"
                : "border-zinc-700 bg-zinc-800 hover:border-zinc-600"
            }`}
          >
            <div className="flex items-center gap-3">
              <div className={`w-4 h-4 rounded-full border-2 flex-shrink-0 ${agentImage === img ? "border-blue-500 bg-blue-500" : "border-zinc-600"}`} />
              <div>
                <p className="font-medium text-zinc-100">{AGENT_IMAGE_INFO[img].label}</p>
                <p className="text-xs text-zinc-500 mt-0.5">{AGENT_IMAGE_INFO[img].desc}</p>
              </div>
            </div>
          </button>
        ))}
      </div>

      <div className="space-y-2">
        <p className="text-sm text-zinc-400">选择 LLM 模型：</p>
        {isLoading ? (
          <div className="text-xs text-zinc-500 py-2">加载模型列表...</div>
        ) : models.length === 0 ? (
          <div className="text-xs text-amber-400 py-2">
            没有可用的{agentImage === "nanoclaw" ? " Anthropic 格式" : ""}模型，请先在 LLM Gateway 中配置 Provider。
          </div>
        ) : (
          <select
            value={model}
            onChange={e => setModel(e.target.value)}
            className="w-full px-3 py-2.5 rounded-lg bg-zinc-800 border border-zinc-700 text-sm text-zinc-100 focus:outline-none focus:border-blue-500 transition-colors"
          >
            <option value="">请选择模型</option>
            {models.map(m => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        )}
        {agentImage === "nanoclaw" && (
          <p className="text-xs text-zinc-600">NanoClaw 仅支持 Anthropic API 格式的模型</p>
        )}
      </div>
    </div>
  );
}
