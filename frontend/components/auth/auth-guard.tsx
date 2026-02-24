"use client";

import { ReactNode, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { getWSClient, destroyWSClient } from "@/lib/ws-singleton";

const PUBLIC_PATHS = ["/login", "/setup"];

export function AuthGuard({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));

    if (isPublic) {
      destroyWSClient();
      return;
    }

    fetch("/api/v1/setup/status")
      .then((r) => r.json())
      .then((data) => {
        if (!data.initialized) {
          router.replace("/setup");
          return;
        }
        const token = localStorage.getItem("lc_token");
        if (!token) {
          router.replace("/login");
          return;
        }
        // 已登录，建立全局 WebSocket 连接
        getWSClient();
      })
      .catch(() => {
        const token = localStorage.getItem("lc_token");
        if (!token) {
          router.replace("/login");
          return;
        }
        getWSClient();
      });
  }, [pathname, router]);

  return <>{children}</>;
}
