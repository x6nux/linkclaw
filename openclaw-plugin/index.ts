/**
 * OpenClaw LinkClaw 插件
 *
 * 将 OpenClaw Agent 连接到 LinkClaw 公司平台，提供：
 * - Channel: 双向消息通道（LinkClaw WS 入站 + WS 出站）
 * - Tools:   17 个 MCP 工具（任务管理、消息、知识库、身份等）
 * - Service: 后台保持 WS 连接 + MCP 心跳
 * - Hook:    自动注入 LinkClaw 身份上下文到系统提示词
 *
 * 安装: 将此目录放到 ~/.openclaw/extensions/linkclaw/
 * 配置: openclaw.yaml → plugins.entries.linkclaw.config.mcpUrl / apiKey
 */
import { MCPClient } from "./src/mcp-client.js";
import { LinkClawBridge } from "./src/channel.js";
import { registerMCPTools } from "./src/tools.js";

// OpenClaw 插件 API 类型（简化声明，避免依赖未发布的 SDK）
interface PluginApi {
  pluginConfig?: Record<string, unknown>;
  logger: { info: (...args: unknown[]) => void; warn: (...args: unknown[]) => void; error: (...args: unknown[]) => void };
  registerTool: (factory: (ctx: unknown) => unknown, opts?: unknown) => void;
  registerHook: (events: string | string[], handler: (event: unknown, ctx: unknown) => unknown, opts?: unknown) => void;
  registerService: (service: { id: string; start: (ctx: unknown) => Promise<void>; stop: (ctx: unknown) => Promise<void> }) => void;
  registerChannel: (registration: { plugin: unknown }) => void;
  registerCommand: (cmd: { name: string; description: string; handler: (ctx: unknown) => Promise<{ text: string }> }) => void;
}

interface PluginConfig {
  mcpUrl: string;
  apiKey: string;
  apiBaseUrl?: string;
}

/**
 * 防御 ReDoS：限制身份缓存最大长度
 */
const MAX_IDENTITY_LENGTH = 10000;

/**
 * 验证 UUID 格式（简单版本，避免复杂正则）
 */
function isValidUUID(s: string): boolean {
  if (s.length !== 36) return false;
  const parts = s.split('-');
  if (parts.length !== 5) return false;
  const [h1, h2, h3, h4, h5] = parts;
  return (
    h1.length === 8 && h2.length === 4 && h3.length === 4 && h4.length === 4 && h5.length === 12 &&
    /^[0-9a-f]+$/i.test(h1) &&
    /^[0-9a-f]+$/i.test(h2) &&
    /^[0-9a-f]+$/i.test(h3) &&
    /^[0-9a-f]+$/i.test(h4) &&
    /^[0-9a-f]+$/i.test(h5)
  );
}

/**
 * 验证 MCP URL 安全性
 * - 仅允许 http/https
 * - 禁止 localhost / 回环 / 内网 / 链路本地地址
 */
function validateMcpUrl(url: string): { valid: boolean; error?: string } {
  try {
    const u = new URL(url);

    if (u.protocol !== "http:" && u.protocol !== "https:") {
      return { valid: false, error: `不支持的协议: ${u.protocol}` };
    }

    const hostname = u.hostname.toLowerCase();

    if (hostname === "localhost" || hostname === "127.0.0.1" || hostname === "[::1]" || hostname === "::1") {
      return { valid: false, error: "禁止本地地址" };
    }

    // 10.0.0.0/8 (RFC 1918)
    if (/^10\./.test(hostname)) {
      return { valid: false, error: "禁止私有 IP: 10.x.x.x" };
    }

    // 172.16.0.0/12 (RFC 1918)
    if (/^172\.(1[6-9]|2[0-9]|3[01])\./.test(hostname)) {
      return { valid: false, error: "禁止私有 IP: 172.16-31.x.x" };
    }

    // 192.168.0.0/16 (RFC 1918)
    if (/^192\.168\./.test(hostname)) {
      return { valid: false, error: "禁止私有 IP: 192.168.x.x" };
    }

    // 127.0.0.0/8 (loopback)
    if (/^127\./.test(hostname)) {
      return { valid: false, error: "禁止回环地址: 127.x.x.x" };
    }

    // 169.254.0.0/16 (link-local)
    if (/^169\.254\./.test(hostname)) {
      return { valid: false, error: "禁止链路本地地址: 169.254.x.x" };
    }

    return { valid: true };
  } catch (e) {
    return { valid: false, error: `URL 格式无效: ${String(e)}` };
  }
}

let mcp: MCPClient | null = null;
let bridge: LinkClawBridge | null = null;
let identityCache = "";

export default {
  id: "linkclaw",
  name: "LinkClaw",
  description: "Connect to LinkClaw company platform via MCP protocol",
  version: "1.0.0",

  async register(api: PluginApi) {
    const cfg = api.pluginConfig as PluginConfig | undefined;
    if (!cfg?.mcpUrl || !cfg?.apiKey) {
      api.logger.error("[LinkClaw] 缺少 mcpUrl 或 apiKey 配置");
      return;
    }

    const validation = validateMcpUrl(cfg.mcpUrl);
    if (!validation.valid) {
      api.logger.error(`[LinkClaw] MCP URL 验证失败: ${validation.error}`);
      return;
    }

    // 验证可选的 apiBaseUrl 配置（SSRF 防护）
    if (cfg.apiBaseUrl) {
      const baseValidation = validateMcpUrl(cfg.apiBaseUrl);
      if (!baseValidation.valid) {
        api.logger.error(`[LinkClaw] API Base URL 验证失败: ${baseValidation.error}`);
        return;
      }
    }

    const log = api.logger;
    mcp = new MCPClient(cfg.mcpUrl, cfg.apiKey);
    bridge = new LinkClawBridge(mcp);

    // ── Service: 后台 WS 连接 + MCP 初始化 ───────────────────

    api.registerService({
      id: "linkclaw-bridge",
      start: async () => {
        log.info("[LinkClaw] 连接 MCP...");
        await mcp!.connect();
        await mcp!.initialize();

        // 加载工具列表
        const tools = await mcp!.listTools();
        log.info(`[LinkClaw] 已加载 ${tools.length} 个 MCP 工具`);

        // 注册所有 MCP 工具为 OpenClaw Agent 工具
        registerMCPTools(api, mcp!, tools);

        // 获取身份信息（缓存用于 hook 注入）
        try {
          identityCache = await mcp!.callTool("get_identity", {});
          // 防御 ReDoS：使用字符串操作代替正则提取 self ID
          if (identityCache.length > MAX_IDENTITY_LENGTH) {
            log.warn("[LinkClaw] 身份信息过长，已截断");
            identityCache = identityCache.slice(0, MAX_IDENTITY_LENGTH);
          }
          // 查找 ID: 或 ID： 后面的 UUID
          const idx1 = identityCache.indexOf("ID:");
          const idx2 = identityCache.indexOf("ID：");
          const idx = idx1 >= 0 ? idx1 : (idx2 >= 0 ? idx2 : -1);
          if (idx >= 0) {
            const start = idx + 3; // "ID:" 或 "ID：" 的长度
            const potentialUUID = identityCache.slice(start, start + 36).trim();
            if (isValidUUID(potentialUUID)) {
              bridge!.setSelfId(potentialUUID);
            } else {
              log.warn("[LinkClaw] 提取的 ID 格式无效:", potentialUUID);
            }
          }
        } catch (e) {
          log.warn("[LinkClaw] 获取身份失败:", e);
        }

        // 启动 WS 事件流 + 心跳
        bridge!.onInbound((msg) => {
          log.info(`[LinkClaw] 收到消息 [${msg.channel}] ${msg.senderId}: ${msg.text.slice(0, 80)}`);
          // 入站消息会被 OpenClaw 的 agent 循环自动处理
          // 通过 channel adapter 的 inbound 机制进入 pipeline
        });
        bridge!.start();
        log.info("[LinkClaw] 事件流已启动");
      },
      stop: async () => {
        bridge?.stop();
        mcp?.disconnect();
        log.info("[LinkClaw] 已断开");
      },
    });

    // ── Channel: LinkClaw 消息通道 ───────────────────────────

    api.registerChannel({
      plugin: {
        id: "linkclaw",
        meta: {
          id: "linkclaw",
          label: "LinkClaw",
          selectionLabel: "LinkClaw (MCP)",
          blurb: "Connect to LinkClaw company platform",
          aliases: ["lc"],
        },
        capabilities: {
          chatTypes: ["direct", "group"],
        },
        outbound: {
          deliveryMode: "direct",
          sendText: async ({ text, recipientId, metadata }: {
            text: string;
            recipientId?: string;
            metadata?: Record<string, unknown>;
          }) => {
            if (!bridge) return { ok: false, error: "not connected" };
            try {
              const channel = metadata?.channel as string | undefined;
              await bridge.sendText({ text, channel, recipientId });
              return { ok: true };
            } catch (e) {
              return { ok: false, error: String(e) };
            }
          },
        },
      },
    });

    // ── Hook: 注入 LinkClaw 身份到系统提示词 ─────────────────

    api.registerHook(
      "before_prompt_build",
      async (event: Record<string, unknown>) => {
        if (!identityCache) return;
        // 将 LinkClaw 身份信息追加到系统提示词
        const parts = (event.systemPromptParts ?? event.parts ?? []) as string[];
        parts.push(
          "\n\n--- LinkClaw Company Context ---\n" + identityCache + "\n--- End LinkClaw Context ---\n"
        );
      },
      { name: "linkclaw.identity-inject", description: "Inject LinkClaw identity into system prompt" },
    );

    // ── Command: 快捷命令 ────────────────────────────────────

    api.registerCommand({
      name: "lc-status",
      description: "Show LinkClaw connection status",
      handler: async () => ({
        text: mcp
          ? `LinkClaw connected to ${cfg.mcpUrl}\nIdentity loaded: ${identityCache ? "yes" : "no"}`
          : "LinkClaw not connected",
      }),
    });

    api.registerCommand({
      name: "lc-tasks",
      description: "List your pending tasks from LinkClaw",
      handler: async () => {
        if (!mcp) return { text: "LinkClaw not connected" };
        try {
          const result = await mcp.callTool("list_tasks", { scope: "mine" });
          return { text: result || "No pending tasks" };
        } catch (e) {
          return { text: `Error: ${e}` };
        }
      },
    });

    log.info("[LinkClaw] 插件已注册");
  },
};
