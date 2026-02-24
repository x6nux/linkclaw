/**
 * WebSocket client with auto-reconnection for LinkClaw channel.
 */
import { WebSocket } from 'ws';
import { logger } from '../logger.js';

export interface WSMessage {
  type: string;
  data: unknown;
}

export interface WSClientOptions {
  url: string;
  onMessage: (msg: WSMessage) => void;
  onConnect?: () => void;
  onDisconnect?: (reason: string) => void;
  reconnectInterval?: number;
}

export class WSClient {
  private opts: WSClientOptions;
  private ws: WebSocket | null = null;
  private _connected = false;
  private shouldReconnect = true;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(opts: WSClientOptions) {
    this.opts = { reconnectInterval: 5000, ...opts };
  }

  connect(): void {
    this.shouldReconnect = true;
    this.doConnect();
  }

  disconnect(): void {
    this.shouldReconnect = false;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.ws?.close();
    this.ws = null;
    this._connected = false;
  }

  isConnected(): boolean {
    return this._connected;
  }

  /** Send a JSON message through the WebSocket */
  send(msg: WSMessage): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      logger.warn('WS send failed: not connected');
      return;
    }
    this.ws.send(JSON.stringify(msg));
  }

  private doConnect(): void {
    // 确保旧连接已关闭
    if (this.ws) {
      this.ws.removeAllListeners();
      this.ws.close();
      this.ws = null;
    }
    this.ws = new WebSocket(this.opts.url);

    this.ws.on('open', () => {
      this._connected = true;
      this.opts.onConnect?.();
    });

    this.ws.on('message', (raw) => {
      try {
        const msg = JSON.parse(raw.toString()) as WSMessage;
        this.opts.onMessage(msg);
      } catch (err) {
        logger.warn({ err, data: raw.toString().slice(0, 100) }, 'Failed to parse WS message');
      }
    });

    this.ws.on('close', () => {
      this._connected = false;
      this.scheduleReconnect('connection closed');
    });

    this.ws.on('error', () => {
      // error 后必定触发 close，由 close 统一处理重连，避免双重调度
      this._connected = false;
    });
  }

  private scheduleReconnect(reason: string): void {
    if (!this.shouldReconnect) return;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.opts.onDisconnect?.(reason);
    logger.info({ reason, interval: this.opts.reconnectInterval }, 'WS will reconnect');
    this.reconnectTimer = setTimeout(() => this.doConnect(), this.opts.reconnectInterval);
  }
}
