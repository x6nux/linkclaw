import { Shell } from "@/components/layout/shell";
import { LLMGatewayContent } from "./content";

export default function LLMPage() {
  return (
    <Shell>
      <div className="max-w-6xl mx-auto space-y-8">
        <div>
          <h1 className="text-xl font-semibold text-zinc-50">LLM 网关</h1>
          <p className="text-sm text-zinc-400 mt-1">
            统一管理 OpenAI / Anthropic API 配置，所有 Agent 通过内部端点调用，支持负载均衡与故障转移
          </p>
        </div>
        <LLMGatewayContent />
      </div>
    </Shell>
  );
}
