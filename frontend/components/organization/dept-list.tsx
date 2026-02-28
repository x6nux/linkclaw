"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Pencil, Plus, Trash2, UserPlus, X } from "lucide-react";
import { toast } from "sonner";
import { useAgents } from "@/hooks/use-agents";
import {
  assignAgent,
  createDepartment,
  deleteDepartment,
  updateDepartment,
  useDepartments,
  type DepartmentPayload,
} from "@/hooks/use-organization";
import type { Department } from "@/lib/types";

interface DeptListProps {
  isChairman: boolean;
}

interface DeptFormState {
  name: string;
  slug: string;
  description: string;
  directorAgentId: string;
  parentDeptId: string;
}

const EMPTY_FORM: DeptFormState = {
  name: "",
  slug: "",
  description: "",
  directorAgentId: "",
  parentDeptId: "",
};

function toFormState(dept?: Department | null): DeptFormState {
  if (!dept) return EMPTY_FORM;
  return {
    name: dept.name,
    slug: dept.slug,
    description: dept.description,
    directorAgentId: dept.director_agent_id ?? "",
    parentDeptId: dept.parent_dept_id ?? "",
  };
}

function toPayload(form: DeptFormState): DepartmentPayload {
  return {
    name: form.name,
    slug: form.slug,
    description: form.description,
    director_agent_id: form.directorAgentId || null,
    parent_dept_id: form.parentDeptId || null,
  };
}

export function DeptList({ isChairman }: DeptListProps) {
  const { departments, isLoading, error } = useDepartments();
  const { agents } = useAgents();

  const [showCreate, setShowCreate] = useState(false);
  const [createForm, setCreateForm] = useState<DeptFormState>(EMPTY_FORM);
  const [editing, setEditing] = useState<Department | null>(null);
  const [editForm, setEditForm] = useState<DeptFormState>(EMPTY_FORM);
  const [assignDept, setAssignDept] = useState<Department | null>(null);
  const [assignAgentId, setAssignAgentId] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const agentNameById = useMemo(
    () => new Map(agents.map((agent) => [agent.id, agent.name])),
    [agents]
  );

  useEffect(() => {
    if (error) toast.error(error.message || "加载部门失败");
  }, [error]);

  function getAgentName(agentId: string) {
    return agentNameById.get(agentId) ?? `${agentId.slice(0, 8)}…`;
  }

  async function handleCreate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSubmitting(true);
    try {
      await createDepartment(toPayload(createForm));
      toast.success("部门已创建");
      setCreateForm(EMPTY_FORM);
      setShowCreate(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "创建部门失败");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleUpdate(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!editing) return;
    setSubmitting(true);
    try {
      await updateDepartment(editing.id, toPayload(editForm));
      toast.success("部门已更新");
      setEditing(null);
      setEditForm(EMPTY_FORM);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "更新部门失败");
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(dept: Department) {
    if (!window.confirm(`确认删除部门「${dept.name}」？`)) return;
    try {
      await deleteDepartment(dept.id);
      toast.success("部门已删除");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "删除部门失败");
    }
  }

  async function handleAssign(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!assignDept) return;
    const agentId = assignAgentId.trim();
    if (!agentId) {
      toast.error("请输入 agent_id");
      return;
    }

    setSubmitting(true);
    try {
      await assignAgent(assignDept.id, agentId);
      toast.success("成员已分配到部门");
      setAssignDept(null);
      setAssignAgentId("");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "分配成员失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg">
      <div className="p-4 border-b border-zinc-800 flex items-center justify-between">
        <div>
          <h2 className="text-zinc-50 text-sm font-semibold">部门管理</h2>
          <p className="text-zinc-400 text-xs mt-1">维护部门、负责人和成员分配</p>
        </div>
        {isChairman && (
          <button
            onClick={() => { setShowCreate((v) => !v); setCreateForm(EMPTY_FORM); }}
            className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-xs transition-colors"
          >
            <Plus className="w-3.5 h-3.5" />
            新增部门
          </button>
        )}
      </div>

      {showCreate && isChairman && (
        <form onSubmit={handleCreate} className="p-4 border-b border-zinc-800 grid grid-cols-1 md:grid-cols-2 gap-3">
          <input value={createForm.name} onChange={(e) => setCreateForm((v) => ({ ...v, name: e.target.value }))}
            className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
            placeholder="部门名称" required />
          <input value={createForm.slug} onChange={(e) => setCreateForm((v) => ({ ...v, slug: e.target.value }))}
            className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
            placeholder="slug（如 engineering）" required />
          <input value={createForm.description} onChange={(e) => setCreateForm((v) => ({ ...v, description: e.target.value }))}
            className="md:col-span-2 bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
            placeholder="部门描述" />
          <select value={createForm.directorAgentId} onChange={(e) => setCreateForm((v) => ({ ...v, directorAgentId: e.target.value }))}
            className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500">
            <option value="">负责人（可选）</option>
            {agents.map((agent) => <option key={agent.id} value={agent.id}>{agent.name}</option>)}
          </select>
          <select value={createForm.parentDeptId} onChange={(e) => setCreateForm((v) => ({ ...v, parentDeptId: e.target.value }))}
            className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500">
            <option value="">父级部门（可选）</option>
            {departments.map((dept) => <option key={dept.id} value={dept.id}>{dept.name}</option>)}
          </select>
          <div className="md:col-span-2 flex justify-end gap-2">
            <button type="button" onClick={() => { setShowCreate(false); setCreateForm(EMPTY_FORM); }}
              className="px-3 py-2 rounded-md text-xs text-zinc-400 hover:text-zinc-200 transition-colors">取消</button>
            <button type="submit" disabled={submitting}
              className="px-3 py-2 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-xs transition-colors disabled:opacity-50">
              {submitting ? "创建中..." : "创建部门"}
            </button>
          </div>
        </form>
      )}

      <div className="overflow-x-auto">
        <table className="min-w-full text-sm">
          <thead>
            <tr className="text-zinc-400 border-b border-zinc-800">
              <th className="text-left font-medium px-4 py-3">名称</th>
              <th className="text-left font-medium px-4 py-3">Slug</th>
              <th className="text-left font-medium px-4 py-3">描述</th>
              <th className="text-left font-medium px-4 py-3">负责人</th>
              {isChairman && <th className="text-right font-medium px-4 py-3">操作</th>}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="border-t border-zinc-800 animate-pulse">
                  <td className="px-4 py-3"><div className="h-4 w-20 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-24 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-48 bg-zinc-800 rounded" /></td>
                  <td className="px-4 py-3"><div className="h-4 w-20 bg-zinc-800 rounded" /></td>
                  {isChairman && <td className="px-4 py-3"><div className="h-4 w-24 bg-zinc-800 rounded ml-auto" /></td>}
                </tr>
              ))
            ) : departments.length === 0 ? (
              <tr>
                <td colSpan={isChairman ? 5 : 4} className="px-4 py-10 text-center text-zinc-500">
                  暂无部门数据
                </td>
              </tr>
            ) : (
              departments.map((dept) => (
                <React.Fragment key={dept.id}>
                  <tr className="border-t border-zinc-800 hover:bg-zinc-950/50 transition-colors">
                    <td className="px-4 py-3 text-zinc-50">{dept.name}</td>
                    <td className="px-4 py-3 text-zinc-400 font-mono">{dept.slug}</td>
                    <td className="px-4 py-3 text-zinc-400">{dept.description || "—"}</td>
                    <td className="px-4 py-3 text-zinc-400">
                      {dept.director_agent_id ? getAgentName(dept.director_agent_id) : "—"}
                    </td>
                    {isChairman && (
                      <td className="px-4 py-3">
                        <div className="flex items-center justify-end gap-2">
                          <button onClick={() => { setEditing(dept); setEditForm(toFormState(dept)); }}
                            className="inline-flex items-center gap-1 px-2 py-1 rounded text-zinc-300 hover:bg-zinc-800 transition-colors">
                            <Pencil className="w-3.5 h-3.5" />编辑
                          </button>
                          <button onClick={() => { setAssignDept(dept); setAssignAgentId(""); }}
                            className="inline-flex items-center gap-1 px-2 py-1 rounded text-zinc-300 hover:bg-zinc-800 transition-colors">
                            <UserPlus className="w-3.5 h-3.5" />分配
                          </button>
                          <button onClick={() => handleDelete(dept)}
                            className="inline-flex items-center gap-1 px-2 py-1 rounded text-red-400 hover:bg-red-500/10 transition-colors">
                            <Trash2 className="w-3.5 h-3.5" />删除
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                </React.Fragment>
              ))
            )}
          </tbody>
        </table>
      </div>

      {editing && (
        <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4">
          <form onSubmit={handleUpdate} className="w-full max-w-xl bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-zinc-50 font-semibold">编辑部门</h3>
              <button type="button" onClick={() => setEditing(null)} className="text-zinc-500 hover:text-zinc-200 transition-colors"><X className="w-4 h-4" /></button>
            </div>
            <input value={editForm.name} onChange={(e) => setEditForm((v) => ({ ...v, name: e.target.value }))}
              className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500" placeholder="部门名称" required />
            <input value={editForm.slug} onChange={(e) => setEditForm((v) => ({ ...v, slug: e.target.value }))}
              className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500" placeholder="slug" required />
            <input value={editForm.description} onChange={(e) => setEditForm((v) => ({ ...v, description: e.target.value }))}
              className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500" placeholder="部门描述" />
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <select value={editForm.directorAgentId} onChange={(e) => setEditForm((v) => ({ ...v, directorAgentId: e.target.value }))}
                className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500">
                <option value="">负责人（可选）</option>
                {agents.map((agent) => <option key={agent.id} value={agent.id}>{agent.name}</option>)}
              </select>
              <select value={editForm.parentDeptId} onChange={(e) => setEditForm((v) => ({ ...v, parentDeptId: e.target.value }))}
                className="bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500">
                <option value="">父级部门（可选）</option>
                {departments.filter((dept) => dept.id !== editing.id).map((dept) => (
                  <option key={dept.id} value={dept.id}>{dept.name}</option>
                ))}
              </select>
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <button type="button" onClick={() => setEditing(null)} className="px-3 py-2 rounded-md text-xs text-zinc-400 hover:text-zinc-200 transition-colors">取消</button>
              <button type="submit" disabled={submitting}
                className="px-3 py-2 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-xs transition-colors disabled:opacity-50">
                {submitting ? "保存中..." : "保存修改"}
              </button>
            </div>
          </form>
        </div>
      )}

      {assignDept && (
        <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4">
          <form onSubmit={handleAssign} className="w-full max-w-md bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-3">
            <div className="flex items-center justify-between">
              <h3 className="text-zinc-50 font-semibold">分配成员到「{assignDept.name}」</h3>
              <button type="button" onClick={() => setAssignDept(null)} className="text-zinc-500 hover:text-zinc-200 transition-colors"><X className="w-4 h-4" /></button>
            </div>
            <input value={assignAgentId} onChange={(e) => setAssignAgentId(e.target.value)}
              className="w-full bg-zinc-950 border border-zinc-800 rounded-md px-3 py-2 text-sm text-zinc-50 focus:outline-none focus:border-blue-500"
              placeholder="输入 agent_id" list="org-agent-options" required />
            <datalist id="org-agent-options">
              {agents.map((agent) => <option key={agent.id} value={agent.id} label={agent.name} />)}
            </datalist>
            <div className="flex justify-end gap-2 pt-1">
              <button type="button" onClick={() => setAssignDept(null)} className="px-3 py-2 rounded-md text-xs text-zinc-400 hover:text-zinc-200 transition-colors">取消</button>
              <button type="submit" disabled={submitting}
                className="px-3 py-2 rounded-md bg-blue-600 hover:bg-blue-500 text-white text-xs transition-colors disabled:opacity-50">
                {submitting ? "提交中..." : "确认分配"}
              </button>
            </div>
          </form>
        </div>
      )}
    </div>
  );
}
