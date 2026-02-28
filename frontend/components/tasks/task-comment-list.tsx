"use client";

import { useState, type FormEvent } from "react";
import { Send, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { addComment, deleteComment } from "@/hooks/use-task-detail";
import type { TaskComment } from "@/lib/types";
import { formatRelativeTime } from "@/lib/utils";

interface TaskCommentListProps {
  taskId: string;
  comments: TaskComment[];
  isLoading: boolean;
}

export function TaskCommentList({ taskId, comments, isLoading }: TaskCommentListProps) {
  const [content, setContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const value = content.trim();
    if (!value) return;

    setSubmitting(true);
    try {
      await addComment(taskId, value);
      setContent("");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (commentId: string) => {
    setDeletingId(commentId);
    try {
      await deleteComment(taskId, commentId);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "操作失败");
    } finally {
      setDeletingId(null);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-2">
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
        <div className="animate-pulse h-4 bg-zinc-800 rounded" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleSubmit} className="space-y-2">
        <textarea
          value={content}
          onChange={(event) => setContent(event.target.value)}
          placeholder="输入评论内容..."
          rows={3}
          className="w-full px-3 py-2 bg-zinc-950 border border-zinc-800 rounded-md text-zinc-50 placeholder-zinc-400 text-sm focus:outline-none focus:border-zinc-700 resize-none"
        />
        <div className="flex justify-end">
          <button
            type="submit"
            disabled={submitting || !content.trim()}
            className="inline-flex items-center gap-1 px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 text-zinc-50 rounded-md text-xs transition-colors"
          >
            <Send className="w-3.5 h-3.5" />
            {submitting ? "提交中..." : "添加评论"}
          </button>
        </div>
      </form>

      <div className="space-y-2">
        {comments.length === 0 ? (
          <p className="text-zinc-400 text-sm">暂无评论</p>
        ) : (
          comments.map((comment) => (
            <div key={comment.id} className="bg-zinc-950 border border-zinc-800 rounded-md p-3 space-y-2">
              <div className="flex items-center justify-between gap-2">
                <p className="text-xs text-zinc-400">
                  {comment.agentId.slice(0, 8)} · {formatRelativeTime(comment.createdAt)}
                </p>
                <button
                  type="button"
                  onClick={() => handleDelete(comment.id)}
                  disabled={deletingId === comment.id}
                  className="inline-flex items-center gap-1 px-2 py-1 text-zinc-400 hover:text-zinc-50 disabled:opacity-50 rounded transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span className="text-xs">
                    {deletingId === comment.id ? "删除中..." : "删除"}
                  </span>
                </button>
              </div>
              <p className="text-sm text-zinc-50 whitespace-pre-wrap break-words">
                {comment.content}
              </p>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
