"use client";

import { useState, useCallback, useEffect, useRef, useMemo } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Shell } from "@/components/layout/shell";
import { ChannelSidebar } from "@/components/messages/channel-sidebar";
import { MessageFeed } from "@/components/messages/message-feed";
import { Composer, ComposerHandle } from "@/components/messages/composer";
import { useAgents } from "@/hooks/use-agents";
import { useMessages } from "@/hooks/use-messages";
import { useRealtime } from "@/hooks/use-realtime";
import { Message, POSITION_LABELS } from "@/lib/types";
import { getWSClient } from "@/lib/ws-singleton";
import { Hash, User } from "lucide-react";

const CHANNEL_IDS = ["general", "engineering", "product", "hr", "random"] as const;

type ActiveTarget =
  | { type: "channel"; id: string; name: string; label: string }
  | { type: "dm"; agentId: string; agentName: string };

interface MsgNewPayload {
  message_id: string;
  company_id: string;
  channel_id?: string;
  channel_name?: string;
  receiver_id?: string;
  sender_id?: string;
  msg_type: string;
  content: string;
  created_at: string;
}

interface ReplyTarget {
  id: string;
  content: string;
  senderLabel: string;
}

export default function MessagesPage() {
  const t = useTranslations("messages");
  const { agents } = useAgents();
  const [currentAgentId, setCurrentAgentId] = useState("");
  const composerRef = useRef<ComposerHandle>(null);
  const [replyTo, setReplyTo] = useState<ReplyTarget | null>(null);

  const channels = CHANNEL_IDS.map((id) => ({
    id,
    name: id,
    label: t(`channel.${id}`),
    isDefault: id === "general",
  }));

  // 从 URL 参数恢复初始聊天目标
  const initialTarget = useMemo((): ActiveTarget => {
    if (typeof window === "undefined") return { type: "channel", id: "general", name: "general", label: t("channel.general") };
    const chat = new URLSearchParams(window.location.search).get("chat");
    if (!chat) return { type: "channel", id: "general", name: "general", label: t("channel.general") };
    if (chat.startsWith("dm:")) {
      const agentId = chat.slice(3);
      return { type: "dm", agentId, agentName: agentId };
    }
    const ch = CHANNEL_IDS.includes(chat as typeof CHANNEL_IDS[number]) ? chat : "general";
    return { type: "channel", id: ch, name: ch, label: t(`channel.${ch}`) };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [active, setActiveState] = useState<ActiveTarget>(initialTarget);

  // 切换目标时同步 URL（不触发导航）
  const setActive = useCallback((target: ActiveTarget) => {
    setActiveState(target);
    const param = target.type === "channel" ? target.name : `dm:${target.agentId}`;
    window.history.replaceState(null, "", `/messages?chat=${param}`);
  }, []);

  // agentMap: id -> 职位-名称（提前计算，供 DM 名称回填使用）
  const agentMap: Record<string, string> = {};
  for (const a of agents) {
    agentMap[a.id] = `${POSITION_LABELS[a.position] ?? a.position}-${a.name}`;
  }
  const agentMapRef = useRef(agentMap);
  agentMapRef.current = agentMap;

  // agents 加载后，回填 DM 目标的显示名称
  useEffect(() => {
    if (active.type === "dm" && active.agentName === active.agentId && agentMap[active.agentId]) {
      setActiveState({ type: "dm", agentId: active.agentId, agentName: agentMap[active.agentId] });
    }
  }, [agents, active, agentMap]);

  useEffect(() => {
    const id = localStorage.getItem("lc_agent_id") ?? "";
    setCurrentAgentId(id);
  }, []);

  const channelName  = active.type === "channel" ? active.name   : undefined;
  const dmReceiverId = active.type === "dm" ? active.agentId : undefined;

  const { messages: fetchedMessages, isLoading, mutate } = useMessages(channelName, dmReceiverId);

  const [extraMessages, setExtraMessages] = useState<Message[]>([]);

  // 切换频道时清空追加消息和回复状态
  useEffect(() => {
    setExtraMessages([]);
    setReplyTo(null);
  }, [active]);

  // WS 实时消息：追加到当前视图 + SWR fallback
  const mutateRef = useRef(mutate);
  mutateRef.current = mutate;

  useRealtime("message.new", useCallback((raw: unknown) => {
    const p = raw as MsgNewPayload;

    let relevant = false;
    if (active.type === "channel") {
      relevant = p.channel_name === active.name;
    } else {
      relevant =
        p.sender_id === active.agentId ||
        p.receiver_id === active.agentId ||
        (p.sender_id === currentAgentId && p.receiver_id === active.agentId);
    }
    if (!relevant) return;

    const realMsg: Message = {
      id: p.message_id,
      company_id: p.company_id,
      sender_id: p.sender_id ?? null,
      channel_id: p.channel_id ?? null,
      receiver_id: p.receiver_id ?? null,
      content: p.content,
      msg_type: (p.msg_type as Message["msg_type"]) ?? "text",
      created_at: p.created_at,
    };

    setExtraMessages((prev) => {
      // 如果真实 id 已存在，跳过
      if (prev.some((m) => m.id === realMsg.id)) return prev;

      // 如果是自己发的消息，替换匹配的乐观消息
      if (p.sender_id === currentAgentId) {
        const optIdx = prev.findIndex(
          (m) => m.id.startsWith("optimistic-") && m.content === p.content
        );
        if (optIdx !== -1) {
          const next = [...prev];
          next[optIdx] = realMsg;
          return next;
        }
      }

      return [...prev, realMsg];
    });

    // Fallback: 重新拉取确保消息一定可见
    mutateRef.current();
  }, [active, currentAgentId]));

  // 消息通知：非当前视图的新消息弹 toast
  useRealtime("message.new", useCallback((raw: unknown) => {
    const p = raw as MsgNewPayload;
    // 跳过自己发的
    if (p.sender_id === currentAgentId) return;

    // 判断是否在当前视图中（已显示则不通知）
    let isCurrentView = false;
    if (active.type === "channel") {
      isCurrentView = p.channel_name === active.name;
    } else {
      isCurrentView = p.sender_id === active.agentId;
    }
    if (isCurrentView) return;

    const senderName = (p.sender_id && agentMapRef.current[p.sender_id]) || "未知";
    const preview = p.content.length > 50 ? p.content.slice(0, 50) + "..." : p.content;
    toast(senderName, { description: preview, duration: 4000 });
  }, [active, currentAgentId]));

  const handleSend = useCallback(async (content: string) => {
    const ws = getWSClient();
    if (active.type === "channel") {
      ws.send("message.send", { channel: active.name, content });
    } else {
      ws.send("message.send", { receiver_id: active.agentId, content });
    }
    const optimistic: Message = {
      id: `optimistic-${Date.now()}`,
      company_id: "",
      sender_id: currentAgentId || null,
      channel_id: null,
      receiver_id: active.type === "dm" ? active.agentId : null,
      content,
      msg_type: "text",
      created_at: new Date().toISOString(),
    };
    setExtraMessages((prev) => [...prev, optimistic]);
  }, [active, currentAgentId]);

  // 合并消息，去除重复，按时间 DESC 排序（MessageFeed 会 reverse 成 ASC 显示）
  const allMessages = (() => {
    const fetchedIds = new Set(fetchedMessages.map((m) => m.id));
    const deduped = extraMessages.filter((e) => {
      // 真实消息已在 fetched 中，跳过
      if (!e.id.startsWith("optimistic-") && fetchedIds.has(e.id)) return false;
      // 乐观消息如果 fetched 已有同内容同发送者，跳过
      if (e.id.startsWith("optimistic-")) {
        return !fetchedMessages.some((f) => f.content === e.content && f.sender_id === e.sender_id);
      }
      return true;
    });
    const merged = [...fetchedMessages, ...deduped];
    merged.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    return merged;
  })();

  // DM 列表过滤掉自己
  const dmTargets = agents
    .filter((a) => a.id !== currentAgentId)
    .map((a) => ({
      id: a.id,
      name: a.name,
      status: a.status,
      position: a.position,
    }));

  const headerTitle = active.type === "channel"
    ? `# ${active.label ?? active.name}`
    : active.agentName;

  const composerPlaceholder = active.type === "channel"
    ? t("sendTo", { channel: active.label ?? active.name })
    : t("dmTo", { name: active.agentName });

  const handleMention = useCallback((senderLabel: string) => {
    composerRef.current?.insertText(`@${senderLabel} `);
  }, []);

  const handleReply = useCallback((msg: Message, senderLabel: string) => {
    setReplyTo({ id: msg.id, content: msg.content, senderLabel });
    composerRef.current?.insertText("");
  }, []);

  return (
    <Shell noPadding>
      <div className="flex h-[calc(100vh-3.5rem)] overflow-hidden">
        <ChannelSidebar
          channels={channels}
          dmTargets={dmTargets}
          activeChannelId={active.type === "channel" ? active.id : undefined}
          activeDMId={active.type === "dm" ? active.agentId : undefined}
          onSelectChannel={(id, name) => {
            const ch = channels.find((c) => c.id === id);
            setActive({ type: "channel", id, name, label: ch?.label ?? name });
          }}
          onSelectDM={(agentId) => {
            setActive({ type: "dm", agentId, agentName: agentMap[agentId] ?? agentId });
          }}
        />

        <div className="flex-1 flex flex-col bg-zinc-950 min-w-0">
          <div className="h-12 flex items-center gap-2 px-4 border-b border-zinc-800 bg-zinc-900 flex-shrink-0">
            {active.type === "channel" ? (
              <Hash className="w-4 h-4 text-zinc-400" />
            ) : (
              <User className="w-4 h-4 text-zinc-400" />
            )}
            <span className="font-medium text-zinc-100 text-sm">{headerTitle}</span>
          </div>

          <MessageFeed
            messages={allMessages}
            agentMap={agentMap}
            currentAgentId={currentAgentId}
            isLoading={isLoading}
            onMention={handleMention}
            onReply={handleReply}
          />

          <Composer
            ref={composerRef}
            placeholder={composerPlaceholder}
            onSend={handleSend}
            replyTo={replyTo}
            onCancelReply={() => setReplyTo(null)}
          />
        </div>
      </div>
    </Shell>
  );
}
