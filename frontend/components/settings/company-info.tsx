"use client";

import useSWR from "swr";
import { api } from "@/lib/api";
import { Building2 } from "lucide-react";

export function CompanyInfo() {
  const { data: company, isLoading } = useSWR(
    "/api/v1/setup/status",
    (url: string) => api.get<{ initialized: boolean; company_slug: string }>(url)
  );

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
      <div className="flex items-center gap-2">
        <Building2 className="w-4 h-4 text-zinc-400" />
        <h2 className="text-sm font-medium text-zinc-200">公司信息</h2>
      </div>
      {isLoading ? (
        <div className="space-y-3 animate-pulse">
          <div>
            <div className="h-3 w-16 bg-zinc-800 rounded mb-2" />
            <div className="h-4 w-32 bg-zinc-800 rounded" />
          </div>
          <div>
            <div className="h-3 w-20 bg-zinc-800 rounded mb-2" />
            <div className="h-4 w-16 bg-zinc-800 rounded" />
          </div>
        </div>
      ) : (
        <div className="space-y-3">
          <div>
            <div className="text-xs text-zinc-500 mb-1">标识 (Slug)</div>
            <div className="text-sm text-zinc-200 font-mono">
              {company?.company_slug || "—"}
            </div>
          </div>
          <div>
            <div className="text-xs text-zinc-500 mb-1">初始化状态</div>
            <div className="text-sm">
              {company?.initialized ? (
                <span className="text-green-400">已初始化</span>
              ) : (
                <span className="text-yellow-400">未初始化</span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
