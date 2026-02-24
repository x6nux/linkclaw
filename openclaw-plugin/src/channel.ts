import { WebSocket } from "ws";
import type { MCPClient } from "./mcp-client.js";

/** LinkClaw 消息事件 payload */
interface MessageNewPayload {
  message_id: string;
  company_id: string;
  channel_id?: string;
  receiver_id?: string;
  sender_id?: string;
  msg_type: string;
  content: string;
  created_at: string;
}

/** LinkClaw 任务事件 payload */
interface TaskPayload {
  task_id: string;
  title: string;
  description?: string;
  assignee_id?: string;
  status?: string;
}

export type InboundHandler = (message: {
  id: string;
  channel: string;
  senderId: string;
  text: string;
  timestamp: string;
  metadata?: Record<string, unknown>;
}) => void;

/**
 * LinkClaw 通道：负责 WebSocket 事件监听和消息收发
 *
 * - inbound: 通过 Agent WS 接收实时事件，转发给 OpenClaw agent 处理
 * - outbound: 通过 WS message.send 帧发送消息
 */
export class LinkClawBridge {
  private ws: WebSocket | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private pingTimer: ReturnType<typeof setInterval> | null = null;
  private handler: InboundHandler | null = null;
  private selfId = "";
  private shouldReconnect = true;

  constructor(private mcp: MCPClient) {}

  /** 设置自身 ID（用于过滤自己发的消息） */
  setSelfId(id: string) { this.selfId = id; }

  /** 注册入站消息处理器 */
  onInbound(handler: InboundHandler) { this.handler = handler; }

  /** 启动 WS 事件流 + MCP 心跳 */
  start() {
    this.shouldReconnect = true;
    this.connectWS();
    this.pingTimer = setInterval(() => {
      this.mcp.ping().catch(() => {});
    }, 30_000);
  }

  /** 发送文本消息到 LinkClaw（通过 WS 帧） */
  async sendText(opts: { text: string; channel?: string; recipientId?: string }) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error("WS not connected");
    }
    const data: Record<string, string> = { content: opts.text };
    if (opts.channel) data.channel = opts.channel;
    if (opts.recipientId) data.receiver_id = opts.recipientId;
    this.ws.send(JSON.stringify({ type: "message.send", data }));
  }

  stop() {
    this.shouldReconnect = false;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    if (this.pingTimer) clearInterval(this.pingTimer);
    this.ws?.close();
    this.ws = null;
  }

  // ── WebSocket 连接 ─────────────────────────────────────────

  private connectWS() {
    const url = this.mcp.agentWSUrl;

    this.ws = new WebSocket(url);

    this.ws.on("open", () => {
      // connected 事件通过 WS 消息推送，不在 open 回调中处理
    });

    this.ws.on("message", (raw) => {
      try {
        const msg = JSON.parse(raw.toString()) as { type: string; data: unknown };
        this.dispatch(msg.type, msg.data);
      } catch { /* ignore parse errors */ }
    });

    this.ws.on("close", () => {
      this.ws = null;
      if (this.shouldReconnect) {
        this.reconnectTimer = setTimeout(() => this.connectWS(), 5000);
      }
    });

    this.ws.on("error", () => {
      // error 后会触发 close，由 close 处理重连
    });
  }

  private dispatch(type: string, payload: unknown) {
    if (!this.handler) return;

    if (type === "message.new") {
      const p = payload as MessageNewPayload;
      if (p.sender_id === this.selfId) return;
      if (p.msg_type !== "text") return;

      const channel = p.receiver_id
        ? `dm:${p.sender_id}`
        : `channel:${p.channel_id ?? "unknown"}`;

      this.handler({
        id: p.message_id,
        channel,
        senderId: p.sender_id ?? "system",
        text: p.content,
        timestamp: p.created_at,
        metadata: { channelId: p.channel_id, receiverId: p.receiver_id },
      });
    }

    if (type === "task.created" || type === "task.updated") {
      const p = payload as TaskPayload;
      const prefix = type === "task.created" ? "[新任务]" : "[任务更新]";
      const text = `${prefix} ${p.title}${p.description ? "\n" + p.description : ""}`;
      this.handler({
        id: `task-${p.task_id}-${Date.now()}`,
        channel: "tasks",
        senderId: "system",
        text,
        timestamp: new Date().toISOString(),
        metadata: { taskId: p.task_id, status: p.status, assigneeId: p.assignee_id },
      });
    }
  }
}
