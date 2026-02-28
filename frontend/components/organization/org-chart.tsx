"use client";

import { useEffect, useMemo } from "react";
import { Users } from "lucide-react";
import { toast } from "sonner";
import { useOrgChart } from "@/hooks/use-organization";
import { cn } from "@/lib/utils";
import type { Agent, OrgDept } from "@/lib/types";

interface DeptNodeProps {
  node: OrgDept;
  depth: number;
  childrenByParent: Map<string, OrgDept[]>;
}

function statusDotClass(status: Agent["status"]) {
  if (status === "online") return "bg-green-500";
  if (status === "busy") return "bg-yellow-500";
  return "bg-zinc-500";
}

function DeptNode({ node, depth, childrenByParent }: DeptNodeProps) {
  const children = childrenByParent.get(node.department.id) ?? [];

  return (
    <div className={cn(depth > 0 && "ml-6 mt-3 pl-4 border-l border-zinc-800")}>
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center justify-between mb-3">
          <div>
            <h3 className="text-zinc-50 font-medium">{node.department.name}</h3>
            <p className="text-zinc-500 text-xs mt-1">
              {node.department.description || "暂无描述"}
            </p>
          </div>
          <div className="inline-flex items-center gap-1.5 text-xs text-zinc-300 bg-zinc-950 border border-zinc-800 rounded-full px-2 py-1">
            <Users className="w-3.5 h-3.5" />
            {node.members.length}
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {node.members.length === 0 ? (
            <span className="text-zinc-500 text-xs">暂无成员</span>
          ) : (
            node.members.map((member) => (
              <span
                key={member.id}
                className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-300 text-xs"
              >
                <span className={cn("w-1.5 h-1.5 rounded-full", statusDotClass(member.status))} />
                {member.name}
              </span>
            ))
          )}
        </div>
      </div>

      {children.map((child) => (
        <DeptNode
          key={child.department.id}
          node={child}
          depth={depth + 1}
          childrenByParent={childrenByParent}
        />
      ))}
    </div>
  );
}

export function OrgChart({ isChairman }: { isChairman?: boolean | null }) {
  const { chart, isLoading, error } = useOrgChart(isChairman === true);

  useEffect(() => {
    if (error) toast.error(error.message || "加载组织架构失败");
  }, [error]);

  const { roots, childrenByParent } = useMemo(() => {
    const departments = chart?.departments ?? [];
    const ids = new Set(departments.map((item) => item.department.id));
    const childMap = new Map<string, OrgDept[]>();
    const rootNodes: OrgDept[] = [];

    for (const item of departments) {
      const parentId = item.department.parent_dept_id;
      if (parentId && ids.has(parentId)) {
        const list = childMap.get(parentId) ?? [];
        list.push(item);
        childMap.set(parentId, list);
      } else {
        rootNodes.push(item);
      }
    }

    for (const list of childMap.values()) {
      list.sort((a, b) => a.department.name.localeCompare(b.department.name));
    }
    rootNodes.sort((a, b) => a.department.name.localeCompare(b.department.name));

    return { roots: rootNodes, childrenByParent: childMap };
  }, [chart]);

  if (isChairman === false) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 text-sm text-zinc-400">
        仅董事长可查看组织架构图。
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 animate-pulse">
            <div className="h-4 w-32 bg-zinc-800 rounded mb-2" />
            <div className="h-3 w-48 bg-zinc-800 rounded mb-4" />
            <div className="flex gap-2">
              <div className="h-6 w-20 bg-zinc-800 rounded" />
              <div className="h-6 w-24 bg-zinc-800 rounded" />
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (roots.length === 0) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-10 text-center text-zinc-500">
        暂无组织架构数据
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {roots.map((root) => (
        <DeptNode
          key={root.department.id}
          node={root}
          depth={0}
          childrenByParent={childrenByParent}
        />
      ))}
    </div>
  );
}
