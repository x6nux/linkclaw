"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Shell } from "@/components/layout/shell";
import { ContextDirectoryList } from "@/components/context/ContextDirectoryList";
import { ContextSearch } from "@/components/context/ContextSearch";
import { ContextDirectoryForm } from "@/components/context/ContextDirectoryForm";
import type { ContextDirectory } from "@/lib/types";

export default function ContextPage() {
  const t = useTranslations("context");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [editingDir, setEditingDir] = useState<ContextDirectory | null>(null);
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  const handleRefresh = () => {
    window.location.reload();
  };

  return (
    <Shell>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-zinc-50">{t("title")}</h1>
            <p className="text-zinc-400 text-sm mt-1">
              {t("subtitle")}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsSearchOpen(true)}
              className="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-sm text-zinc-200 transition-colors flex items-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              {t("search")}
            </button>
            <button
              onClick={() => setIsCreateOpen(true)}
              className="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors"
            >
              {t("addDirectory")}
            </button>
          </div>
        </div>

        {/* Directory List */}
        <ContextDirectoryList
          onEdit={setEditingDir}
          onRefresh={handleRefresh}
        />
      </div>

      {/* Create Dialog */}
      {isCreateOpen && (
        <ContextDirectoryForm
          onClose={() => setIsCreateOpen(false)}
          onSubmitSuccess={() => {
            setIsCreateOpen(false);
            handleRefresh();
          }}
        />
      )}

      {/* Edit Dialog */}
      {editingDir && (
        <ContextDirectoryForm
          directory={editingDir}
          onClose={() => setEditingDir(null)}
          onSubmitSuccess={() => {
            setEditingDir(null);
            handleRefresh();
          }}
        />
      )}

      {/* Search Modal */}
      {isSearchOpen && (
        <ContextSearch
          onClose={() => setIsSearchOpen(false)}
        />
      )}
    </Shell>
  );
}
