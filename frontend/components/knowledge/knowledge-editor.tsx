"use client";

import { useState, useEffect } from "react";
import ReactMarkdown from "react-markdown";
import { toast } from "sonner";
import { Save, Trash2, Eye, Edit3 } from "lucide-react";
import { cn } from "@/lib/utils";
import { KnowledgeDoc, updateDoc, deleteDoc, createDoc } from "@/hooks/use-knowledge";

interface Props {
  doc: KnowledgeDoc | null;
  isNew?: boolean;
  onSaved: (doc: KnowledgeDoc) => void;
  onDeleted: () => void;
}

export function KnowledgeEditor({ doc, isNew, onSaved, onDeleted }: Props) {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [tagsInput, setTagsInput] = useState("");
  const [mode, setMode] = useState<"edit" | "preview">("edit");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (isNew) {
      setTitle("");
      setContent("");
      setTagsInput("");
      setMode("edit");
    } else if (doc) {
      setTitle(doc.title);
      setContent(doc.content);
      setTagsInput(doc.tags.join(", "));
    }
  }, [doc, isNew]);

  const tags = tagsInput
    .split(",")
    .map((t) => t.trim())
    .filter(Boolean);

  const handleSave = async () => {
    if (!title.trim()) return;
    setSaving(true);
    try {
      let result: KnowledgeDoc;
      if (isNew || !doc) {
        result = await createDoc(title, content, tags);
      } else {
        result = await updateDoc(doc.id, title, content, tags);
      }
      toast.success("文档已保存");
      onSaved(result);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!doc) return;
    toast("确认删除这篇文档？", {
      action: {
        label: "删除",
        onClick: async () => {
          try {
            await deleteDoc(doc.id);
            toast.success("文档已删除");
            onDeleted();
          } catch {
            toast.error("删除失败");
          }
        },
      },
      cancel: { label: "取消", onClick: () => {} },
    });
  };

  if (!doc && !isNew) {
    return (
      <div className="flex-1 flex items-center justify-center text-zinc-500 text-sm">
        请从左侧选择一篇文档
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-4 py-2 border-b border-zinc-800 flex-shrink-0">
        <div className="flex items-center gap-1 bg-zinc-800 rounded-md p-0.5">
          <button
            onClick={() => setMode("edit")}
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors",
              mode === "edit" ? "bg-zinc-700 text-zinc-50" : "text-zinc-400 hover:text-zinc-200"
            )}
          >
            <Edit3 className="w-3 h-3" /> 编辑
          </button>
          <button
            onClick={() => setMode("preview")}
            className={cn(
              "flex items-center gap-1 px-2 py-1 rounded text-xs transition-colors",
              mode === "preview" ? "bg-zinc-700 text-zinc-50" : "text-zinc-400 hover:text-zinc-200"
            )}
          >
            <Eye className="w-3 h-3" /> 预览
          </button>
        </div>
        <div className="flex-1" />
        <button
          onClick={handleSave}
          disabled={saving || !title.trim()}
          className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-xs font-medium transition-colors"
        >
          <Save className="w-3 h-3" />
          {saving ? "保存中..." : "保存"}
        </button>
        {!isNew && doc && (
          <button
            onClick={handleDelete}
            className="flex items-center gap-1 px-3 py-1.5 bg-red-600/20 hover:bg-red-600/40 text-red-400 rounded-md text-xs font-medium transition-colors"
          >
            <Trash2 className="w-3 h-3" />
            删除
          </button>
        )}
      </div>

      {/* Meta fields */}
      <div className="px-4 py-3 border-b border-zinc-800 space-y-2 flex-shrink-0">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="文档标题"
          className="w-full px-0 py-1 bg-transparent border-none text-zinc-50 text-lg font-semibold placeholder-zinc-600 focus:outline-none"
        />
        <input
          value={tagsInput}
          onChange={(e) => setTagsInput(e.target.value)}
          placeholder="标签（逗号分隔）"
          className="w-full px-0 py-0.5 bg-transparent border-none text-zinc-400 text-sm placeholder-zinc-600 focus:outline-none"
        />
      </div>

      {/* Content */}
      <div className="flex-1 overflow-hidden">
        {mode === "edit" ? (
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="使用 Markdown 编写文档内容..."
            className="w-full h-full p-4 bg-transparent text-zinc-200 text-sm font-mono placeholder-zinc-600 focus:outline-none resize-none"
          />
        ) : (
          <div className="h-full overflow-y-auto p-4 prose prose-invert prose-sm max-w-none">
            <ReactMarkdown>{content || "*暂无内容*"}</ReactMarkdown>
          </div>
        )}
      </div>
    </div>
  );
}
