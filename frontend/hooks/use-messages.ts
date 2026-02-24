"use client";

import useSWR from "swr";
import { api } from "@/lib/api";
import { Message, PaginatedResponse } from "@/lib/types";

// 首次连接一次性拉取全部历史（limit=500）
const INITIAL_LIMIT = 500;

export function useMessages(channel?: string, receiverId?: string, beforeId?: string) {
  const query = new URLSearchParams();
  if (channel)    query.set("channel", channel);
  if (receiverId) query.set("receiver_id", receiverId);
  if (beforeId)   query.set("before_id", beforeId);
  query.set("limit", String(INITIAL_LIMIT));

  const key = (channel || receiverId)
    ? `/api/v1/messages?${query.toString()}`
    : null;

  const { data, error, isLoading, mutate } = useSWR(
    key,
    (url) => api.get<PaginatedResponse<Message>>(url),
    {
      revalidateOnFocus: false,    // 切回窗口不重新拉取（WS 保持实时）
      revalidateOnReconnect: true, // 断线重连后重新拉取
    }
  );
  return {
    messages: data?.data ?? [],
    cursor: data?.cursor,
    isLoading,
    error,
    mutate,
  };
}
