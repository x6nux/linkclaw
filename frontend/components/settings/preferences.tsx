"use client";

import { useTheme } from "next-themes";
import { Sun, Moon, Globe } from "lucide-react";
import { cn } from "@/lib/utils";

export function Preferences() {
  const { theme, setTheme } = useTheme();
  const locale =
    typeof window !== "undefined"
      ? localStorage.getItem("lc_locale") || "zh"
      : "zh";

  const switchLocale = (l: string) => {
    localStorage.setItem("lc_locale", l);
    window.location.reload();
  };

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-6">
      <div className="flex items-center gap-2">
        <h2 className="text-sm font-medium text-zinc-200">偏好设置</h2>
      </div>

      {/* Theme */}
      <div className="space-y-2">
        <div className="text-xs text-zinc-500">主题</div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setTheme("dark")}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md text-sm border transition-colors",
              theme === "dark"
                ? "border-blue-500 bg-blue-500/10 text-blue-400"
                : "border-zinc-700 text-zinc-400 hover:border-zinc-600"
            )}
          >
            <Moon className="w-4 h-4" /> 深色
          </button>
          <button
            onClick={() => setTheme("light")}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md text-sm border transition-colors",
              theme === "light"
                ? "border-blue-500 bg-blue-500/10 text-blue-400"
                : "border-zinc-700 text-zinc-400 hover:border-zinc-600"
            )}
          >
            <Sun className="w-4 h-4" /> 浅色
          </button>
        </div>
      </div>

      {/* Language */}
      <div className="space-y-2">
        <div className="text-xs text-zinc-500">语言</div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => switchLocale("zh")}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md text-sm border transition-colors",
              locale === "zh"
                ? "border-blue-500 bg-blue-500/10 text-blue-400"
                : "border-zinc-700 text-zinc-400 hover:border-zinc-600"
            )}
          >
            <Globe className="w-4 h-4" /> 中文
          </button>
          <button
            onClick={() => switchLocale("en")}
            className={cn(
              "flex items-center gap-2 px-3 py-2 rounded-md text-sm border transition-colors",
              locale === "en"
                ? "border-blue-500 bg-blue-500/10 text-blue-400"
                : "border-zinc-700 text-zinc-400 hover:border-zinc-600"
            )}
          >
            <Globe className="w-4 h-4" /> English
          </button>
        </div>
      </div>
    </div>
  );
}
