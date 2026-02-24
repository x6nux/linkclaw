"use client";

import { useEffect, useRef } from "react";
import { Message } from "@/lib/types";
import { formatRelativeTime } from "@/lib/utils";
import { TaskProgressCard } from "./task-progress-card";
import { MarkdownContent } from "./markdown-content";

interface MessageFeedProps {
  messages: Message[];
  currentAgentId?: string;
  agentMap?: Record<string, string>;
  isLoading?: boolean;
  onMention?: (senderLabel: string) => void;
  onReply?: (msg: Message, senderLabel: string) => void;
}

function MessageBubble({
  msg,
  isSelf,
  senderLabel,
  onMention,
  onReply,
}: {
  msg: Message;
  isSelf: boolean;
  senderLabel: string;
  onMention?: (senderLabel: string) => void;
  onReply?: (msg: Message, senderLabel: string) => void;
}) {
  if (msg.msgType === "task_update" && msg.taskMeta) {
    return (
      <div className="flex gap-3 group">
        <div className="w-7 h-7 rounded-full bg-blue-500/20 flex items-center justify-center flex-shrink-0 mt-0.5">
          <span className="text-blue-400 text-xs">{"\u2699"}</span>
        </div>
        <div className="flex-1">
          <span className="text-xs text-zinc-500 mb-1 block">{"系统"} · {formatRelativeTime(msg.createdAt)}</span>
          <TaskProgressCard meta={msg.taskMeta} />
        </div>
      </div>
    );
  }

  const avatarChar = senderLabel[0]?.toUpperCase() ?? "?";

  return (
    <div className={`flex gap-3 group ${isSelf ? "flex-row-reverse" : ""}`}>
      <div
        className="w-7 h-7 rounded-full bg-zinc-700 flex items-center justify-center flex-shrink-0 mt-0.5 cursor-pointer select-none"
        onDoubleClick={() => onMention?.(senderLabel)}
        title="双击 @提及"
      >
        <span className="text-zinc-300 text-xs font-medium">{avatarChar}</span>
      </div>

      <div className={`flex flex-col max-w-[80%] ${isSelf ? "items-end" : "items-start"}`}>
        <div className="flex items-baseline gap-2 mb-0.5">
          {!isSelf && (
            <span className="text-xs font-medium text-zinc-300">{senderLabel}</span>
          )}
          <span className="text-xs text-zinc-600">{formatRelativeTime(msg.createdAt)}</span>
        </div>
        <div
          className={`px-3 py-2 rounded-2xl text-sm leading-relaxed select-text ${
            isSelf
              ? "bg-blue-600 text-white rounded-tr-sm"
              : "bg-zinc-800 text-zinc-100 rounded-tl-sm"
          }`}
          onDoubleClick={() => onReply?.(msg, senderLabel)}
          title="双击回复"
        >
          <MarkdownContent content={msg.content} />
        </div>
      </div>
    </div>
  );
}

export function MessageFeed({ messages, currentAgentId, agentMap, isLoading, onMention, onReply }: MessageFeedProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  if (isLoading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-zinc-500 text-sm">加载中...</div>
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <p className="text-zinc-400 text-sm">暂无消息</p>
          <p className="text-zinc-600 text-xs mt-1">发送第一条消息开始对话</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
      {[...messages].reverse().map((msg) => {
        const isSelf = !!currentAgentId && msg.senderId === currentAgentId;
        const senderLabel =
          (msg.senderId && agentMap?.[msg.senderId]) ||
          msg.senderName ||
          "未知";
        return (
          <MessageBubble
            key={msg.id}
            msg={msg}
            isSelf={isSelf}
            senderLabel={senderLabel}
            onMention={onMention}
            onReply={onReply}
          />
        );
      })}
      <div ref={bottomRef} />
    </div>
  );
}
