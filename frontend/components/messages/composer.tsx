"use client";

import { useState, useRef, useImperativeHandle, forwardRef, KeyboardEvent } from "react";
import { Send, X } from "lucide-react";
import { cn } from "@/lib/utils";

export interface ComposerHandle {
  insertText: (text: string) => void;
}

interface ReplyTarget {
  id: string;
  content: string;
  senderLabel: string;
}

interface ComposerProps {
  placeholder?: string;
  onSend: (content: string) => Promise<void>;
  disabled?: boolean;
  replyTo?: ReplyTarget | null;
  onCancelReply?: () => void;
}

export const Composer = forwardRef<ComposerHandle, ComposerProps>(
  function Composer({ placeholder = "输入消息...", onSend, disabled, replyTo, onCancelReply }, ref) {
    const [content, setContent] = useState("");
    const [sending, setSending] = useState(false);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useImperativeHandle(ref, () => ({
      insertText(text: string) {
        const el = textareaRef.current;
        if (!el) {
          setContent((prev) => prev + text);
          return;
        }
        const start = el.selectionStart;
        const end = el.selectionEnd;
        const before = content.slice(0, start);
        const after = content.slice(end);
        const next = before + text + after;
        setContent(next);
        requestAnimationFrame(() => {
          el.focus();
          const cursor = start + text.length;
          el.setSelectionRange(cursor, cursor);
        });
      },
    }));

    const handleSend = async () => {
      const text = content.trim();
      if (!text || sending) return;
      setSending(true);
      setContent("");

      let finalContent = text;
      if (replyTo) {
        const quoted = replyTo.content.split("\n").map((l) => `> ${l}`).join("\n");
        finalContent = `> **${replyTo.senderLabel}**:\n${quoted}\n\n${text}`;
      }

      try {
        await onSend(finalContent);
      } finally {
        setSending(false);
        onCancelReply?.();
        textareaRef.current?.focus();
      }
    };

    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if (e.key === "Enter" && !e.shiftKey) {
        e.preventDefault();
        handleSend();
      }
    };

    return (
      <div className="border-t border-zinc-800 p-3 bg-zinc-900">
        {replyTo && (
          <div className="flex items-center gap-2 mb-2 px-3 py-1.5 bg-zinc-800 rounded-lg text-xs">
            <span className="text-zinc-400 flex-1 truncate">
              回复 <span className="text-zinc-200 font-medium">{replyTo.senderLabel}</span>
              ：{replyTo.content.slice(0, 60)}{replyTo.content.length > 60 ? "..." : ""}
            </span>
            <button onClick={onCancelReply} className="text-zinc-500 hover:text-zinc-300 flex-shrink-0">
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}
        <div className="flex items-end gap-2 bg-zinc-800 rounded-xl px-3 py-2">
          <textarea
            ref={textareaRef}
            value={content}
            onChange={(e) => setContent(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={placeholder}
            disabled={disabled || sending}
            rows={1}
            className={cn(
              "flex-1 bg-transparent text-zinc-100 text-sm placeholder-zinc-500 resize-none outline-none",
              "min-h-[24px] max-h-32 leading-6",
              (disabled || sending) && "opacity-50 cursor-not-allowed"
            )}
            style={{ height: "auto" }}
            onInput={(e) => {
              const el = e.currentTarget;
              el.style.height = "auto";
              el.style.height = `${Math.min(el.scrollHeight, 128)}px`;
            }}
          />
          <button
            onClick={handleSend}
            disabled={!content.trim() || disabled || sending}
            className={cn(
              "p-1.5 rounded-lg transition-colors flex-shrink-0",
              content.trim() && !disabled && !sending
                ? "text-blue-400 hover:text-blue-300 hover:bg-blue-500/10"
                : "text-zinc-600 cursor-not-allowed"
            )}
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
        <p className="text-xs text-zinc-600 mt-1 px-1">Enter 发送 · Shift+Enter 换行</p>
      </div>
    );
  }
);
