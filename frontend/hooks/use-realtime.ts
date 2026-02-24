"use client";

import { useEffect, useRef } from "react";
import { getWSClient } from "@/lib/ws-singleton";

export function useRealtime(
  event: string,
  handler: (data: unknown) => void
) {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    const client = getWSClient();
    const unsubscribe = client.on(event, (data) => handlerRef.current(data));
    return () => {
      unsubscribe();
    };
  }, [event]);
}
