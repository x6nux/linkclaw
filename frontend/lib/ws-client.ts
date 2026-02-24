type EventHandler = (data: unknown) => void;

const HEARTBEAT_INTERVAL = 30_000; // 30s
const RECONNECT_BASE = 1_000;      // 1s
const RECONNECT_MAX = 30_000;      // 30s

export class WSClient {
  private ws: WebSocket | null = null;
  private handlers = new Map<string, Set<EventHandler>>();
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectAttempts = 0;
  private url: string;

  constructor(url: string) {
    this.url = url;
  }

  connect() {
    this.clearReconnectTimer();
    const token = localStorage.getItem("lc_token");
    const wsUrl = `${this.url}?token=${token ?? ""}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.send("ping", {});
      this.startHeartbeat();
      this.emit("connected", {});
    };
    this.ws.onclose = () => {
      this.stopHeartbeat();
      this.emit("disconnected", {});
      this.scheduleReconnect();
    };
    this.ws.onerror = (err) => this.emit("error", err);
    this.ws.onmessage = (e) => {
      let parsed: { type: string; data: unknown };
      try {
        parsed = JSON.parse(e.data as string);
      } catch {
        return; // ignore malformed JSON
      }
      this.emit(parsed.type, parsed.data);
    };
  }

  disconnect() {
    this.clearReconnectTimer();
    this.stopHeartbeat();
    this.ws?.close();
    this.ws = null;
  }

  on(event: string, handler: EventHandler): () => void {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(handler);
    return () => {
      this.handlers.get(event)?.delete(handler);
    };
  }

  send(type: string, data: unknown) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, data }));
    }
  }

  private emit(event: string, data: unknown) {
    this.handlers.get(event)?.forEach((h) => h(data));
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      this.send("ping", {});
    }, HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
  }

  private scheduleReconnect() {
    const delay = Math.min(RECONNECT_BASE * 2 ** this.reconnectAttempts, RECONNECT_MAX);
    this.reconnectAttempts++;
    this.reconnectTimer = setTimeout(() => this.connect(), delay);
  }
}
