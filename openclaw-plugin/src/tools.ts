import type { MCPClient, MCPTool } from "./mcp-client.js";

/**
 * 将 LinkClaw MCP 工具注册为 OpenClaw Agent 工具
 *
 * OpenClaw 的 registerTool 接受一个工厂函数，返回工具定义
 * 每个工具都通过 MCPClient.callTool 代理到 LinkClaw 后端
 */
export function registerMCPTools(
  api: { registerTool: (factory: (ctx: unknown) => unknown, opts?: unknown) => void },
  mcp: MCPClient,
  mcpTools: MCPTool[],
) {
  for (const tool of mcpTools) {
    const toolDef = buildToolDef(tool, mcp);
    api.registerTool(
      () => toolDef,
      { name: `linkclaw.${tool.name}`, optional: true },
    );
  }
}

function buildToolDef(tool: MCPTool, mcp: MCPClient) {
  return {
    name: `linkclaw_${tool.name}`,
    description: `[LinkClaw] ${tool.description}`,
    inputSchema: {
      type: "object" as const,
      properties: convertProps(tool.inputSchema.properties),
      required: tool.inputSchema.required,
    },
    execute: async (params: Record<string, unknown>) => {
      try {
        const result = await mcp.callTool(tool.name, params);
        return { text: result };
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        return { text: `[LinkClaw Error] ${msg}`, isError: true };
      }
    },
  };
}

function convertProps(
  props?: Record<string, { type: string; description: string; enum?: string[] }>,
): Record<string, unknown> {
  if (!props) return {};
  const out: Record<string, unknown> = {};
  for (const [k, v] of Object.entries(props)) {
    const p: Record<string, unknown> = { type: v.type, description: v.description };
    if (v.enum) p.enum = v.enum;
    out[k] = p;
  }
  return out;
}
