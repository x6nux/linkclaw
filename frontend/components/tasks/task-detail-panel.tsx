"use client";

import { useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { Eye, EyeOff, MessageSquare, Pencil, Tag, Workflow, X } from "lucide-react";
import { toast } from "sonner";
import {
  addWatcher,
  removeWatcher,
  updateTags,
  useTaskDetail,
} from "@/hooks/use-task-detail";
import { TaskCommentList } from "@/components/tasks/task-comment-list";
import { TaskDependencyList } from "@/components/tasks/task-dependency-list";
import { cn, formatDate, getPriorityColor, getStatusColor } from "@/lib/utils";

type DetailTab = "comments" | "dependencies";

interface TaskDetailPanelProps {
  taskId: string;
}

export function TaskDetailPanel({ taskId }: TaskDetailPanelProps) {
  const { task, isLoading, error } = useTaskDetail(taskId);
  const [activeTab, setActiveTab] = useState<DetailTab>("comments");
  const [isEditingTags, setIsEditingTags] = useState(false);
  const [draftTags, setDraftTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState("");
  const [savingTags, setSavingTags] = useState(false);
  const [watching, setWatching] = useState(false);
  const [currentAgentId, setCurrentAgentId] = useState<string | null>(null);

  useEffect(() => {
    setCurrentAgentId(localStorage.getItem("lc_agent_id"));
  }, []);

  useEffect(() => {
    if (!task) return;
    setDraftTags(task.tags ?? []);
  }, [task]);

  useEffect(() => {
    if (!error) return;
    toast.error(error instanceof Error ? error.message : "操作失败");
  }, [error]);

  const watcherIds = useMemo(
    () => (task?.watchers ?? []).map((watcher) => watcher.agent_id),
    [task]
  );

  const isWatching = !!currentAgentId && watcherIds.includes(currentAgentId);

  const handleTagKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key !== "Enter" && event.key !== ",") return;
    event.preventDefault();

    const nextTag = tagInput.trim().replace(/,$/, "");
    if (!nextTag) return;

    setDraftTags((prev) => (prev.includes(nextTag) ? prev : [...prev, nextTag]));
    setTagInput("");
  };

  const handleRemoveTag = (tag: string) => {
    setDraftTags((prev) => prev.filter((item) => item !== tag));
  };

  const handleSaveTags = async () => {
    if (!task) return;

    const bufferedTag = tagInput.trim().replace(/,$/, "");
    const nextTags = Array.from(
      new Set(
        [...draftTags, ...(bufferedTag ? [bufferedTag] : [])]
          .map((item) => item.trim())
          .filter(Boolean)
      )
    );

    setSavingTags(true);
    try {
      await updateTags(task.id, nextTags);
      setDraftTags(nextTags);
      setTagInput("");
      setIsEditingTags(false);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSavingTags(false);
    }
  };

  const handleCancelTags = () => {
    setDraftTags(task?.tags ?? []);
    setTagInput("");
    setIsEditingTags(false);
  };

  const handleToggleWatch = async () => {
    if (!task || !currentAgentId) return;

    setWatching(true);
    try {
      if (isWatching) {
        await removeWatcher(task.id);
      } else {
        await addWatcher(task.id);
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setWatching(false);
    }
  };

  if (isLoading && !task) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-3">
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
      </div>
    );
  }

  if (!task) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5">
        <p className="text-zinc-400 text-sm">任务不存在或无权限访问</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-5">
        <header className="space-y-2">
          <div className="flex items-start justify-between gap-3">
            <div>
              <h1 className="text-xl font-semibold text-zinc-50">{task.title}</h1>
              <p className="text-xs text-zinc-400 mt-1">截止时间：{formatDate(task.due_at)}</p>
            </div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 px-2 py-1 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-50 text-xs">
                <span className={cn("w-2 h-2 rounded-full", getStatusColor(task.status))} />
                {task.status}
              </span>
              <span
                className={cn(
                  "px-2 py-1 rounded-md bg-zinc-950 border border-zinc-800 text-xs",
                  getPriorityColor(task.priority)
                )}
              >
                {task.priority}
              </span>
            </div>
          </div>
        </header>

        <section className="space-y-2">
          <h2 className="text-sm font-medium text-zinc-50">描述</h2>
          <p className="text-sm text-zinc-400 whitespace-pre-wrap break-words">
            {task.description || "暂无描述"}
          </p>
        </section>

        <section className="space-y-2">
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-zinc-50 inline-flex items-center gap-1.5">
              <Tag className="w-4 h-4 text-zinc-400" />
              标签
            </h2>
            {!isEditingTags && (
              <button
                type="button"
                onClick={() => setIsEditingTags(true)}
                className="inline-flex items-center gap-1 px-2 py-1 text-zinc-400 hover:text-zinc-50 text-xs transition-colors"
              >
                <Pencil className="w-3.5 h-3.5" />
                编辑
              </button>
            )}
          </div>

          {isEditingTags ? (
            <div className="space-y-2">
              <div className="flex flex-wrap gap-2">
                {draftTags.map((tag) => (
                  <span key={tag} className="inline-flex items-center gap-1 px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-xs text-zinc-50">
                    {tag}
                    <button type="button" onClick={() => handleRemoveTag(tag)} className="text-zinc-400 hover:text-zinc-50">
                      <X className="w-3 h-3" />
                    </button>
                  </span>
                ))}
              </div>
              <input
                value={tagInput}
                onChange={(event) => setTagInput(event.target.value)}
                onKeyDown={handleTagKeyDown}
                placeholder="输入标签，按 Enter 或逗号添加"
                className="w-full px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-md text-zinc-50 placeholder-zinc-400 text-sm focus:outline-none focus:border-zinc-700"
              />
              <div className="flex gap-2 justify-end">
                <button type="button" onClick={handleCancelTags} className="px-3 py-1.5 rounded-md text-xs text-zinc-400 hover:text-zinc-50 transition-colors">取消</button>
                <button
                  type="button"
                  onClick={handleSaveTags}
                  disabled={savingTags}
                  className="px-3 py-1.5 rounded-md text-xs bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 text-zinc-50 transition-colors"
                >
                  {savingTags ? "保存中..." : "保存标签"}
                </button>
              </div>
            </div>
          ) : (
            <div className="flex flex-wrap gap-2">
              {(task.tags ?? []).length === 0 ? (
                <p className="text-sm text-zinc-400">暂无标签</p>
              ) : (
                task.tags.map((tag) => (
                  <span key={tag} className="px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-xs text-zinc-50">{tag}</span>
                ))
              )}
            </div>
          )}
        </section>

        <section className="space-y-2">
          <div className="flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-zinc-50 inline-flex items-center gap-1.5">
              <Eye className="w-4 h-4 text-zinc-400" />
              关注者
            </h2>
            <button
              type="button"
              onClick={handleToggleWatch}
              disabled={!currentAgentId || watching}
              className="inline-flex items-center gap-1 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 text-zinc-50 rounded-md text-xs transition-colors"
            >
              {isWatching ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
              {isWatching ? "取消关注" : "关注任务"}
            </button>
          </div>
          {watcherIds.length === 0 ? (
            <p className="text-sm text-zinc-400">暂无关注者</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {watcherIds.map((agentId) => (
                <span key={agentId} className="px-2 py-1 rounded bg-zinc-950 border border-zinc-800 text-xs text-zinc-50">{agentId}</span>
              ))}
            </div>
          )}
        </section>
      </div>

      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-5 space-y-4">
        <div className="flex items-center gap-2 border-b border-zinc-800 pb-2">
          <button
            type="button"
            onClick={() => setActiveTab("comments")}
            className={cn("inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-sm transition-colors", activeTab === "comments" ? "bg-zinc-800 text-zinc-50" : "text-zinc-400 hover:text-zinc-50")}
          >
            <MessageSquare className="w-4 h-4" />
            评论
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("dependencies")}
            className={cn("inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-sm transition-colors", activeTab === "dependencies" ? "bg-zinc-800 text-zinc-50" : "text-zinc-400 hover:text-zinc-50")}
          >
            <Workflow className="w-4 h-4" />
            依赖任务
          </button>
        </div>

        {activeTab === "comments" ? (
          <TaskCommentList taskId={task.id} comments={task.comments ?? []} isLoading={isLoading} />
        ) : (
          <TaskDependencyList taskId={task.id} dependencies={task.dependencies ?? []} isLoading={isLoading} />
        )}
      </div>
    </div>
  );
}
