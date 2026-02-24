/**
 * LinkClaw Channel — connects nanoclaw to the LinkClaw backend.
 *
 * Inbound: Agent WebSocket (/api/v1/agents/me/ws) for message events.
 * Outbound: WebSocket frames (message.send) for sending messages.
 * MCP: Configured separately in container settings.json (see container-runner.ts).
 */
import { logger } from '../logger.js';
import { Channel, NewMessage, OnChatMetadata, OnInboundMessage } from '../types.js';
import { WSClient, WSMessage } from './ws-client.js';

export interface LinkClawChannelOptions {
  baseUrl: string;
  apiKey: string;
  onMessage: OnInboundMessage;
  onChatMetadata: OnChatMetadata;
  onInitRequired?: (prompt: string) => void;
}

/** JID: lc:ch:<channel_name> for channels, lc:dm:<agent_id> for DMs */
function channelJid(name: string): string {
  return `lc:ch:${name}`;
}
function dmJid(agentId: string): string {
  return `lc:dm:${agentId}`;
}

function parseJid(jid: string): { channel: string } | { receiverId: string } | null {
  if (jid.startsWith('lc:ch:')) return { channel: jid.slice(6) };
  if (jid.startsWith('lc:dm:')) return { receiverId: jid.slice(6) };
  return null;
}

// Backend event payloads
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

interface AgentInfo {
  id: string;
  name: string;
}

interface ChannelInfo {
  id: string;
  name: string;
}

export class LinkClawChannel implements Channel {
  readonly name = 'linkclaw';

  private opts: LinkClawChannelOptions;
  private wsClient: WSClient | null = null;
  private _connected = false;
  private agentId = '';
  private initialized = false;

  // Caches: id → name
  private agentNames = new Map<string, string>();
  private channelNames = new Map<string, string>();

  constructor(opts: LinkClawChannelOptions) {
    this.opts = opts;
  }

  async connect(): Promise<void> {
    await Promise.all([this.fetchAgentNames(), this.fetchChannelNames()]);

    // Convert http(s) baseUrl to ws(s) for WebSocket
    const wsUrl = this.opts.baseUrl
      .replace(/^https:/, 'wss:')
      .replace(/^http:/, 'ws:');

    this.wsClient = new WSClient({
      url: `${wsUrl}/api/v1/agents/me/ws?token=${this.opts.apiKey}`,
      onMessage: (msg) => this.handleMessage(msg),
      onConnect: () => {
        this._connected = true;
        logger.info('LinkClaw WS connected');
      },
      onDisconnect: (reason) => {
        this._connected = false;
        logger.warn({ reason }, 'LinkClaw WS disconnected');
      },
    });
    this.wsClient.connect();
  }

  async sendMessage(jid: string, text: string): Promise<void> {
    const parsed = parseJid(jid);
    if (!parsed) throw new Error(`Invalid LinkClaw JID: ${jid}`);

    if (!this.wsClient?.isConnected()) {
      throw new Error('WS not connected');
    }

    const data: Record<string, string> = { content: text };
    if ('channel' in parsed) {
      data.channel = parsed.channel;
    } else {
      data.receiver_id = parsed.receiverId;
    }

    this.wsClient.send({ type: 'message.send', data });
  }

  isConnected(): boolean {
    return this._connected;
  }

  ownsJid(jid: string): boolean {
    return jid.startsWith('lc:');
  }

  async disconnect(): Promise<void> {
    this.wsClient?.disconnect();
    this._connected = false;
  }

  // --- Event handling ---

  private handleMessage(msg: WSMessage): void {
    const { type, data } = msg;

    if (type === 'connected') {
      const d = data as Record<string, unknown>;
      this.agentId = (d.agent_id as string) || '';
      this.initialized = d.initialized === true;
      logger.info({ agentId: this.agentId, initialized: this.initialized }, 'LinkClaw agent connected');
      return;
    }

    if (type === 'init_required') {
      const d = data as Record<string, string>;
      if (d.prompt && this.opts.onInitRequired) {
        logger.info('Received init_required, triggering initialization');
        this.opts.onInitRequired(d.prompt);
      }
      return;
    }

    logger.info({ eventType: type }, 'WS event received');

    if (type === 'message.new') {
      this.handleMessageNew(data as MessageNewPayload);
    }
    // task.created / task.updated — could be handled here in the future
  }

  private handleMessageNew(p: MessageNewPayload): void {
    logger.info({ messageId: p.message_id, senderId: p.sender_id, channelId: p.channel_id, receiverId: p.receiver_id }, 'handleMessageNew received');
    // Skip own messages to avoid echo loops
    if (p.sender_id === this.agentId) {
      logger.debug({ messageId: p.message_id }, 'Skipping own message');
      return;
    }
    // Skip non-text messages
    if (p.msg_type !== 'text') {
      logger.debug({ messageId: p.message_id, msgType: p.msg_type }, 'Skipping non-text message');
      return;
    }

    // Resolve JID: channel messages use channel name, DMs use sender id
    let chatJid: string;
    let isGroup: boolean;

    if (p.channel_id) {
      const chName = this.channelNames.get(p.channel_id) || p.channel_id;
      chatJid = channelJid(chName);
      isGroup = true;
    } else if (p.sender_id) {
      chatJid = dmJid(p.sender_id);
      isGroup = false;
    } else {
      return;
    }

    const senderName = (p.sender_id && this.agentNames.get(p.sender_id)) || p.sender_id || 'system';

    const newMsg: NewMessage = {
      id: p.message_id,
      chat_jid: chatJid,
      sender: p.sender_id || 'system',
      sender_name: senderName,
      content: p.content,
      timestamp: p.created_at,
      is_from_me: false,
    };

    this.opts.onChatMetadata(chatJid, p.created_at, undefined, 'linkclaw', isGroup);
    this.opts.onMessage(chatJid, newMsg);
  }

  // --- API helpers ---

  private async fetchAgentNames(): Promise<void> {
    try {
      const res = await fetch(`${this.opts.baseUrl}/api/v1/agents`, {
        headers: { Authorization: `Bearer ${this.opts.apiKey}` },
      });
      if (!res.ok) return;
      const json = (await res.json()) as { data?: AgentInfo[] };
      for (const a of json.data || []) {
        if (a.id && a.name) this.agentNames.set(a.id, a.name);
      }
      logger.debug({ count: this.agentNames.size }, 'Agent names cached');
    } catch (err) {
      logger.warn({ err }, 'Failed to fetch agent names');
    }
  }

  private async fetchChannelNames(): Promise<void> {
    try {
      const res = await fetch(`${this.opts.baseUrl}/api/v1/messages/channels`, {
        headers: { Authorization: `Bearer ${this.opts.apiKey}` },
      });
      if (!res.ok) return;
      const json = (await res.json()) as { data?: ChannelInfo[] };
      for (const c of json.data || []) {
        if (c.id && c.name) this.channelNames.set(c.id, c.name);
      }
      logger.debug({ count: this.channelNames.size }, 'Channel names cached');
    } catch (err) {
      logger.warn({ err }, 'Failed to fetch channel names');
    }
  }
}
