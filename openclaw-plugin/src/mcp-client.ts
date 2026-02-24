import { EventSource } from "eventsource";

/** MCP 工具定义 */
export interface MCPTool {
  name: string;
  description: string;
  inputSchema: {
    type: string;
    properties?: Record<string, { type: string; description: string; enum?: string[] }>;
    required?: string[];
  };
}

interface JSONRPCResponse {
  jsonrpc: string;
  id: number | string | null;
  result?: unknown;
  error?: { code: number; message: string };
}

type PendingRequest = {
  resolve: (value: unknown) => void;
  reject: (reason: Error) => void;
};

/**
 * LinkClaw MCP SSE 客户端
 *
 * MCP 通道: GET /mcp/sse (Bearer auth) → endpoint 事件（含 session_id）
 *           POST /mcp/message?session_id=xxx → JSON-RPC 请求，响应通过 SSE 回传
 * Agent 事件: WS /api/v1/agents/me/ws?token=xxx → 双向消息通道
 */
export class MCPClient {
  private es: EventSource | null = null;
  private endpoint = "";
  private reqId = 0;
  private pending = new Map<number | string, PendingRequest>();
  private _ready: Promise<void>;
  private _resolveReady!: () => void;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private baseUrl: string;
  private mcpUrl: string;
  private apiKey: string;

  constructor(mcpUrl: string, apiKey: string) {
    this.mcpUrl = mcpUrl;
    this.apiKey = apiKey;
    this.baseUrl = mcpUrl.replace(/\/mcp\/sse$/, "");
    this._ready = new Promise((r) => { this._resolveReady = r; });
  }

  /** 建立 SSE 连接并等待 endpoint */
  connect(): Promise<void> {
    this.setupSSE();
    return this._ready;
  }

  private setupSSE() {
    const token = this.apiKey;
    this.es = new EventSource(this.mcpUrl, {
      fetch: (input, init) =>
        fetch(input, {
          ...init,
          headers: { ...(init?.headers ?? {}), Authorization: `Bearer ${token}` },
        }),
    });

    this.es.addEventListener("endpoint", (evt) => {
      const path = (evt as MessageEvent).data as string;
      this.endpoint = `${this.baseUrl}${path}`;
      this._resolveReady();
    });

    this.es.addEventListener("message", (evt) => {
      try {
        const resp = JSON.parse((evt as MessageEvent).data) as JSONRPCResponse;
        const p = this.pending.get(resp.id!);
        if (!p) return;
        this.pending.delete(resp.id!);
        resp.error
          ? p.reject(new Error(`RPC ${resp.error.code}: ${resp.error.message}`))
          : p.resolve(resp.result);
      } catch { /* keepalive 等非 JSON 数据 */ }
    });

    this.es.onerror = () => {
      this.es?.close();
      this.es = null;
      this._ready = new Promise((r) => { this._resolveReady = r; });
      this.reconnectTimer = setTimeout(() => this.setupSSE(), 5000);
    };
  }

  /** 发送 JSON-RPC 请求 */
  async request(method: string, params?: unknown): Promise<unknown> {
    await this._ready;
    const id = ++this.reqId;

    const promise = new Promise<unknown>((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      setTimeout(() => {
        if (this.pending.has(id)) {
          this.pending.delete(id);
          reject(new Error(`MCP timeout: ${method}`));
        }
      }, 30_000);
    });

    const res = await fetch(this.endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ jsonrpc: "2.0", id, method, params: params ?? {} }),
    });
    if (!res.ok) {
      this.pending.delete(id);
      throw new Error(`MCP POST ${res.status}`);
    }
    return promise;
  }

  async initialize(): Promise<void> {
    await this.request("initialize", {
      protocolVersion: "2024-11-05",
      clientInfo: { name: "openclaw-linkclaw", version: "1.0.0" },
      capabilities: {},
    });
  }

  async listTools(): Promise<MCPTool[]> {
    const r = (await this.request("tools/list")) as { tools: MCPTool[] };
    return r.tools;
  }

  async callTool(name: string, args: Record<string, unknown>): Promise<string> {
    const r = (await this.request("tools/call", { name, arguments: args })) as {
      content: Array<{ type: string; text: string }>;
      isError?: boolean;
    };
    const text = r.content.map((b) => b.text).join("\n");
    if (r.isError) throw new Error(text);
    return text;
  }

  async ping(): Promise<void> { await this.request("ping"); }

  /** Agent WebSocket URL（事件流） */
  get agentWSUrl(): string {
    const wsBase = this.baseUrl
      .replace(/^https:/, "wss:")
      .replace(/^http:/, "ws:");
    return `${wsBase}/api/v1/agents/me/ws?token=${this.apiKey}`;
  }

  get token(): string { return this.apiKey; }

  disconnect() {
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.es?.close();
    this.es = null;
    for (const [, p] of this.pending) p.reject(new Error("disconnected"));
    this.pending.clear();
  }
}
