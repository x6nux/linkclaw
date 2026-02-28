"use client";

import { useState, type ChangeEvent, type DragEvent, type FormEvent } from "react";
import { Loader2, Upload, X } from "lucide-react";
import { mutate } from "swr";
import { toast } from "sonner";
import { useAgents } from "@/hooks/use-agents";
import { useTasks } from "@/hooks/use-tasks";
import type { TaskPriority } from "@/lib/types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";
const MAX_FILE_SIZE = 10 * 1024 * 1024;

const PRIORITY_OPTIONS: TaskPriority[] = ["low", "medium", "high", "urgent"];
const ALLOWED_EXTENSIONS = new Set([
  ".jpg",
  ".jpeg",
  ".png",
  ".gif",
  ".webp",
  ".bmp",
  ".svg",
  ".pdf",
  ".doc",
  ".docx",
  ".xls",
  ".xlsx",
  ".ppt",
  ".pptx",
  ".txt",
  ".md",
  ".csv",
  ".zip",
  ".tar.gz",
]);
const ALLOWED_MIME_TYPES = new Set([
  "application/pdf",
  "application/msword",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-powerpoint",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
  "text/plain",
  "text/markdown",
  "text/csv",
  "application/zip",
  "application/x-zip-compressed",
  "application/x-tar",
  "application/gzip",
  "application/x-gzip",
]);

const FILE_ACCEPT_ATTR = [
  "image/*",
  ".pdf",
  ".doc",
  ".docx",
  ".xls",
  ".xlsx",
  ".ppt",
  ".pptx",
  ".txt",
  ".md",
  ".csv",
  ".zip",
  ".tar.gz",
].join(",");

function getFileKey(file: File): string {
  return `${file.name}-${file.size}-${file.lastModified}`;
}

function getFileExtension(name: string): string {
  const lower = name.toLowerCase().trim();
  if (lower.endsWith(".tar.gz")) return ".tar.gz";
  const dot = lower.lastIndexOf(".");
  return dot >= 0 ? lower.slice(dot) : "";
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function validateFile(file: File): string | null {
  if (file.size > MAX_FILE_SIZE) {
    return `${file.name} exceeds 10MB`;
  }

  const mimeType = file.type.toLowerCase().trim();
  if (mimeType.startsWith("image/")) {
    return null;
  }
  if (ALLOWED_MIME_TYPES.has(mimeType)) {
    return null;
  }

  const ext = getFileExtension(file.name);
  if (!ALLOWED_EXTENSIONS.has(ext)) {
    return `${file.name} is not an allowed file type`;
  }
  return null;
}

function distributeProgress(files: File[], loadedBytes: number): Record<string, number> {
  if (files.length === 0) return {};

  const progressByFile: Record<string, number> = {};
  let remaining = loadedBytes;

  for (const file of files) {
    const key = getFileKey(file);
    if (file.size <= 0) {
      progressByFile[key] = 100;
      continue;
    }

    const consumed = Math.min(Math.max(remaining, 0), file.size);
    progressByFile[key] = Math.round((consumed / file.size) * 100);
    remaining -= consumed;
  }

  return progressByFile;
}

async function submitTask(
  formData: FormData,
  files: File[],
  setProgress: (value: Record<string, number>) => void
) {
  await new Promise<void>((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${API_BASE}/api/v1/tasks`);

    const token = localStorage.getItem("lc_token");
    if (token) {
      xhr.setRequestHeader("Authorization", `Bearer ${token}`);
    }

    const totalFileBytes = files.reduce((sum, file) => sum + Math.max(file.size, 0), 0);
    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable || totalFileBytes <= 0) return;
      const loaded = Math.min(event.loaded, totalFileBytes);
      setProgress(distributeProgress(files, loaded));
    };

    xhr.onerror = () => {
      reject(new Error("Network error"));
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        setProgress(
          Object.fromEntries(files.map((file) => [getFileKey(file), 100]))
        );
        resolve();
        return;
      }

      let errorMessage = `Create task failed (${xhr.status})`;
      try {
        const parsed = JSON.parse(xhr.responseText) as { error?: string };
        if (parsed?.error) {
          errorMessage = parsed.error;
        }
      } catch {
        // no-op
      }
      reject(new Error(errorMessage));
    };

    xhr.send(formData);
  });
}

export function TaskCreateForm() {
  const { agents } = useAgents();
  const { tasks } = useTasks();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<TaskPriority>("medium");
  const [assigneeID, setAssigneeID] = useState("");
  const [parentID, setParentID] = useState("");
  const [files, setFiles] = useState<File[]>([]);
  const [fileProgress, setFileProgress] = useState<Record<string, number>>({});
  const [dragging, setDragging] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  const addFiles = (incoming: File[]) => {
    if (incoming.length === 0) return;

    setFiles((prev) => {
      const existing = new Set(prev.map(getFileKey));
      const next = [...prev];

      for (const file of incoming) {
        const error = validateFile(file);
        if (error) {
          toast.error(error);
          continue;
        }

        const key = getFileKey(file);
        if (existing.has(key)) continue;
        existing.add(key);
        next.push(file);
      }

      return next;
    });
  };

  const handleFileInput = (event: ChangeEvent<HTMLInputElement>) => {
    addFiles(Array.from(event.target.files ?? []));
    event.target.value = "";
  };

  const handleDrop = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setDragging(false);
    addFiles(Array.from(event.dataTransfer.files ?? []));
  };

  const handleDragOver = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    event.stopPropagation();
  };

  const handleDragEnter = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setDragging(true);
  };

  const handleDragLeave = (event: DragEvent<HTMLLabelElement>) => {
    event.preventDefault();
    event.stopPropagation();
    setDragging(false);
  };

  const removeFile = (key: string) => {
    setFiles((prev) => prev.filter((file) => getFileKey(file) !== key));
    setFileProgress((prev) => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      toast.error("Title is required");
      return;
    }

    const formData = new FormData();
    formData.append("title", trimmedTitle);
    formData.append("description", description.trim());
    formData.append("priority", priority);
    if (assigneeID) formData.append("assignee_id", assigneeID);
    if (parentID) formData.append("parent_id", parentID);
    files.forEach((file) => formData.append("files", file, file.name));

    setSubmitting(true);
    setFileProgress(Object.fromEntries(files.map((file) => [getFileKey(file), 0])));

    try {
      await submitTask(formData, files, setFileProgress);
      await mutate((key) => typeof key === "string" && key.startsWith("/api/v1/tasks"));
      toast.success("Task created");

      setTitle("");
      setDescription("");
      setPriority("medium");
      setAssigneeID("");
      setParentID("");
      setFiles([]);
      setFileProgress({});
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Create task failed");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="bg-zinc-900 border border-zinc-800 rounded-lg p-4 space-y-4"
    >
      <div className="flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold text-zinc-100">Create Task</h2>
        {submitting && (
          <span className="inline-flex items-center gap-1 text-xs text-zinc-400">
            <Loader2 className="w-3.5 h-3.5 animate-spin" />
            Uploading...
          </span>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        <input
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          placeholder="Title"
          required
          className="px-3 py-2 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-zinc-700"
        />
        <select
          value={priority}
          onChange={(event) => setPriority(event.target.value as TaskPriority)}
          className="px-3 py-2 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-100 text-sm focus:outline-none focus:border-zinc-700"
        >
          {PRIORITY_OPTIONS.map((value) => (
            <option key={value} value={value}>
              {value}
            </option>
          ))}
        </select>
        <select
          value={assigneeID}
          onChange={(event) => setAssigneeID(event.target.value)}
          className="px-3 py-2 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-100 text-sm focus:outline-none focus:border-zinc-700"
        >
          <option value="">Assignee (optional)</option>
          {agents.map((agent) => (
            <option key={agent.id} value={agent.id}>
              {agent.name}
            </option>
          ))}
        </select>
        <select
          value={parentID}
          onChange={(event) => setParentID(event.target.value)}
          className="px-3 py-2 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-100 text-sm focus:outline-none focus:border-zinc-700"
        >
          <option value="">Parent task (optional)</option>
          {tasks.map((task) => (
            <option key={task.id} value={task.id}>
              {task.title}
            </option>
          ))}
        </select>
      </div>

      <textarea
        value={description}
        onChange={(event) => setDescription(event.target.value)}
        placeholder="Description"
        rows={3}
        className="w-full px-3 py-2 rounded-md bg-zinc-950 border border-zinc-800 text-zinc-100 placeholder-zinc-500 text-sm focus:outline-none focus:border-zinc-700"
      />

      <input
        id="task-create-files"
        type="file"
        multiple
        accept={FILE_ACCEPT_ATTR}
        onChange={handleFileInput}
        className="sr-only"
      />
      <label
        htmlFor="task-create-files"
        onDrop={handleDrop}
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={[
          "flex flex-col items-center justify-center gap-2 rounded-md border border-dashed px-4 py-6 text-sm transition-colors cursor-pointer",
          dragging
            ? "border-blue-500/60 bg-blue-500/10 text-blue-300"
            : "border-zinc-700 bg-zinc-950 text-zinc-400 hover:text-zinc-200 hover:border-zinc-600",
        ].join(" ")}
      >
        <Upload className="w-5 h-5" />
        <p>Drag and drop files here, or click to select</p>
        <p className="text-xs text-zinc-500">
          Max 10MB each. Images, documents, and archives (.zip, .tar.gz)
        </p>
      </label>

      <div className="space-y-2">
        {files.length === 0 ? (
          <p className="text-xs text-zinc-500">No files selected</p>
        ) : (
          files.map((file) => {
            const key = getFileKey(file);
            const progress = fileProgress[key] ?? 0;
            return (
              <div
                key={key}
                className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 space-y-1"
              >
                <div className="flex items-center justify-between gap-2">
                  <div className="min-w-0">
                    <p className="text-sm text-zinc-200 truncate">{file.name}</p>
                    <p className="text-xs text-zinc-500">{formatFileSize(file.size)}</p>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeFile(key)}
                    disabled={submitting}
                    className="p-1 rounded text-zinc-500 hover:text-zinc-200 hover:bg-zinc-800 disabled:opacity-50"
                    aria-label={`Remove ${file.name}`}
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
                <div className="w-full h-1.5 bg-zinc-800 rounded">
                  <div
                    className="h-1.5 bg-blue-500 rounded transition-all"
                    style={{ width: `${progress}%` }}
                  />
                </div>
              </div>
            );
          })
        )}
      </div>

      <div className="flex justify-end">
        <button
          type="submit"
          disabled={submitting}
          className="inline-flex items-center gap-2 px-4 py-2 rounded-md bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium transition-colors"
        >
          {submitting ? <Loader2 className="w-4 h-4 animate-spin" /> : null}
          Create Task
        </button>
      </div>
    </form>
  );
}
