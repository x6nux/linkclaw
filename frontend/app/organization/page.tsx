"use client";

import { useEffect, useState } from "react";
import { Shell } from "@/components/layout/shell";
import { ApprovalList } from "@/components/organization/approval-list";
import { DeptList } from "@/components/organization/dept-list";
import { OrgChart } from "@/components/organization/org-chart";
import type { Agent } from "@/lib/types";

type TabKey = "chart" | "departments" | "approvals";

const TABS: { key: TabKey; label: string }[] = [
  { key: "chart", label: "组织架构图" },
  { key: "departments", label: "部门管理" },
  { key: "approvals", label: "审批请求" },
];

export default function OrganizationPage() {
  const [activeTab, setActiveTab] = useState<TabKey>("chart");
  const [isChairman, setIsChairman] = useState<boolean | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("lc_token");
    if (!token) return;
    fetch("/api/v1/agents", { headers: { Authorization: `Bearer ${token}` } })
      .then((r) => r.json())
      .then((d) => {
        const agentId = localStorage.getItem("lc_agent_id");
        const me = (d.data as Agent[])?.find((a) => a.id === agentId);
        if (me?.role_type === "chairman") setIsChairman(true);
      })
      .catch(() => {});
  }, []);

  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">组织架构</h1>
          <p className="text-zinc-400 text-sm mt-1">查看组织结构、管理部门并处理审批流程</p>
        </div>

        <div className="inline-flex items-center bg-zinc-900 border border-zinc-800 rounded-lg p-1">
          {TABS.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`px-3 py-1.5 rounded-md text-sm transition-colors ${
                activeTab === tab.key
                  ? "bg-blue-500/10 text-blue-400"
                  : "text-zinc-400 hover:text-zinc-200"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <div>
          {activeTab === "chart" && <OrgChart isChairman={isChairman} />}
          {activeTab === "departments" && <DeptList isChairman={isChairman ?? false} />}
          {activeTab === "approvals" && <ApprovalList isChairman={isChairman ?? false} />}
        </div>
      </div>
    </Shell>
  );
}
