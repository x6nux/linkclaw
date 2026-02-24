"use client";

import { cn } from "@/lib/utils";
import { POSITION_LABELS } from "@/lib/types";
import { Hash, ChevronDown, ChevronRight } from "lucide-react";
import { useState } from "react";

interface Channel {
  id: string;
  name: string;
  label?: string; // 显示名（i18n）
  isDefault?: boolean;
}

interface DMTarget {
  id: string;
  name: string;
  status: "online" | "busy" | "offline";
  position: string;
}

interface ChannelSidebarProps {
  channels: Channel[];
  dmTargets: DMTarget[];
  activeChannelId?: string;
  activeDMId?: string;
  onSelectChannel: (channelId: string, channelName: string) => void;
  onSelectDM: (agentId: string) => void;
}

function StatusDot({ status }: { status: string }) {
  const colors: Record<string, string> = {
    online: "bg-green-500",
    busy: "bg-yellow-500",
    offline: "bg-zinc-500",
  };
  return (
    <span
      className={cn(
        "w-2 h-2 rounded-full flex-shrink-0",
        colors[status] ?? "bg-zinc-500"
      )}
    />
  );
}

export function ChannelSidebar({
  channels,
  dmTargets,
  activeChannelId,
  activeDMId,
  onSelectChannel,
  onSelectDM,
}: ChannelSidebarProps) {
  const [channelsOpen, setChannelsOpen] = useState(true);
  const [dmsOpen, setDmsOpen] = useState(true);

  return (
    <div className="w-56 flex-shrink-0 bg-zinc-900 border-r border-zinc-800 flex flex-col overflow-y-auto">
      {/* 频道分组 */}
      <div className="mt-4">
        <button
          onClick={() => setChannelsOpen((v) => !v)}
          className="flex items-center gap-1 px-3 py-1 w-full text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider hover:text-zinc-200 transition-colors"
        >
          {channelsOpen ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          频道
        </button>
        {channelsOpen && (
          <div className="mt-1 space-y-0.5 px-2">
            {channels.map((ch) => (
              <button
                key={ch.id}
                onClick={() => onSelectChannel(ch.id, ch.name)}
                className={cn(
                  "flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm transition-colors",
                  activeChannelId === ch.id
                    ? "bg-zinc-700 text-zinc-50"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800"
                )}
              >
                <Hash className="w-3.5 h-3.5 flex-shrink-0" />
                <span className="truncate">{ch.label ?? ch.name}</span>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* DM 分组 */}
      <div className="mt-4">
        <button
          onClick={() => setDmsOpen((v) => !v)}
          className="flex items-center gap-1 px-3 py-1 w-full text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider hover:text-zinc-200 transition-colors"
        >
          {dmsOpen ? (
            <ChevronDown className="w-3 h-3" />
          ) : (
            <ChevronRight className="w-3 h-3" />
          )}
          直接消息
        </button>
        {dmsOpen && (
          <div className="mt-1 space-y-0.5 px-2">
            {dmTargets.length === 0 && (
              <p className="text-xs text-zinc-600 px-2 py-1">暂无同事</p>
            )}
            {dmTargets.map((agent) => (
              <button
                key={agent.id}
                onClick={() => onSelectDM(agent.id)}
                className={cn(
                  "flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm transition-colors",
                  activeDMId === agent.id
                    ? "bg-zinc-700 text-zinc-50"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800"
                )}
              >
                <StatusDot status={agent.status} />
                <span className="truncate">
                  {POSITION_LABELS[agent.position] ?? agent.position}-{agent.name}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
