"use client";

import { useState, useEffect } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { ContextDirectory } from "@/lib/types";
import { cn } from "@/lib/utils";

interface Props {
  directory?: ContextDirectory;
  onClose: () => void;
  onSubmitSuccess: () => void;
}

export function ContextDirectoryForm({ directory, onClose, onSubmitSuccess }: Props) {
  const t = useTranslations("context");
  const isEdit = !!directory;

  const [name, setName] = useState("");
  const [path, setPath] = useState("");
  const [description, setDescription] = useState("");
  const [isActive, setIsActive] = useState(true);
  const [filePatterns, setFilePatterns] = useState("");
  const [excludePatterns, setExcludePatterns] = useState("");
  const [maxFileSize, setMaxFileSize] = useState(1024);
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    if (directory) {
      setName(directory.name);
      setPath(directory.path);
      setDescription(directory.description || "");
      setIsActive(directory.is_active);
      setFilePatterns(directory.file_patterns || "");
      setExcludePatterns(directory.exclude_patterns || "");
      setMaxFileSize(directory.max_file_size);
    }
  }, [directory]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      const token = localStorage.getItem("lc_token");
      const body = {
        name,
        path,
        description,
        is_active: isActive,
        file_patterns: filePatterns,
        exclude_patterns: excludePatterns,
        max_file_size: maxFileSize,
      };

      const url = isEdit
        ? `/api/v1/context/directories/${directory.id}`
        : "/api/v1/context/directories";

      const res = await fetch(url, {
        method: isEdit ? "PUT" : "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: "操作失败" }));
        throw new Error(err.error || "操作失败");
      }

      toast.success(isEdit ? t("editSuccess") : t("createSuccess"));
      onSubmitSuccess();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("errors.operationFailed"));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-lg mx-4 shadow-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="border-b border-zinc-800 px-5 pt-5 pb-3">
          <h2 className="text-lg font-semibold text-zinc-50">
            {isEdit ? t("editDirectory") : t("addDirectory")}
          </h2>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              {t("form.name")} <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm focus:outline-none focus:border-blue-500 transition-colors"
              placeholder={t("form.namePlaceholder")}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              {t("form.path")} <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={path}
              onChange={(e) => setPath(e.target.value)}
              required
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors"
              placeholder={t("form.pathPlaceholder")}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              {t("form.description")}
            </label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm focus:outline-none focus:border-blue-500 transition-colors resize-none"
              placeholder={t("form.optional")}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">
                {t("form.filePatterns")}
              </label>
              <input
                type="text"
                value={filePatterns}
                onChange={(e) => setFilePatterns(e.target.value)}
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors"
                placeholder={t("form.filePatternsPlaceholder")}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-zinc-300 mb-1">
                {t("form.excludePatterns")}
              </label>
              <input
                type="text"
                value={excludePatterns}
                onChange={(e) => setExcludePatterns(e.target.value)}
                className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm font-mono focus:outline-none focus:border-blue-500 transition-colors"
                placeholder={t("form.excludePatternsPlaceholder")}
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-zinc-300 mb-1">
              {t("form.maxFileSize")}
            </label>
            <input
              type="number"
              value={maxFileSize}
              onChange={(e) => setMaxFileSize(Number(e.target.value))}
              min={1}
              className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-zinc-50 text-sm focus:outline-none focus:border-blue-500 transition-colors"
            />
          </div>

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="is_active"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
              className="w-4 h-4 rounded bg-zinc-800 border-zinc-700 text-blue-600 focus:ring-blue-500"
            />
            <label htmlFor="is_active" className="text-sm text-zinc-300">
              {t("form.isActive")}
            </label>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-2 pt-4 border-t border-zinc-800">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-sm text-zinc-300 transition-colors"
            >
              {t("common.cancel")}
            </button>
            <button
              type="submit"
              disabled={isSubmitting}
              className={cn(
                "px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors disabled:opacity-40"
              )}
            >
              {isSubmitting ? t("form.submitting") : isEdit ? t("common.save") : t("common.create")}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
