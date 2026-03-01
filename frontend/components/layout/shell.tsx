"use client";

import { ReactNode, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { destroyWSClient } from "@/lib/ws-singleton";
import { useTranslations } from "next-intl";
import {
  LayoutDashboard,
  Bot,
  CheckSquare,
  MessageSquare,
  BookOpen,
  Settings,
  Menu,
  X,
  Zap,
  Cpu,
  Brain,
  ScrollText,
  Sun,
  Moon,
  Globe,
  LogOut,
  Building2,
  Activity,
  Database,
} from "lucide-react";
import { useRouter } from "next/navigation";

function Sidebar({ collapsed }: { collapsed: boolean }) {
  const t = useTranslations();
  const pathname = usePathname();

  const nav = [
    { href: "/dashboard", label: t("nav.dashboard"), icon: LayoutDashboard },
    { href: "/agents", label: t("nav.agents"), icon: Bot },
    { href: "/prompts", label: t("nav.prompts"), icon: ScrollText },
    { href: "/tasks", label: t("nav.tasks"), icon: CheckSquare },
    { href: "/messages", label: t("nav.messages"), icon: MessageSquare },
    { href: "/knowledge", label: t("nav.knowledge"), icon: BookOpen },
    { href: "/memories", label: t("nav.memories"), icon: Brain },
    { href: "/context", label: t("nav.context"), icon: Database },
    { href: "/llm", label: t("nav.llm"), icon: Cpu },
    { href: "/organization", label: t("nav.organization"), icon: Building2 },
    { href: "/observability", label: t("nav.observability"), icon: Activity },
    { href: "/settings", label: t("nav.settings"), icon: Settings },
  ];

  return (
    <aside
      className={cn(
        "fixed left-0 top-0 h-full bg-zinc-900 border-r border-zinc-800/50 flex flex-col transition-all duration-300 ease-in-out z-40 overflow-hidden",
        collapsed ? "w-16" : "w-60"
      )}
    >
      <div className={cn(
        "flex items-center gap-2 px-4 h-14 border-b border-zinc-800/50 transition-colors",
        !collapsed && "hover:border-zinc-700/50"
      )}>
        <div className="w-5 h-5 rounded-md bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center flex-shrink-0 shadow-lg shadow-blue-500/20">
          <Zap className="w-3 h-3 text-white" />
        </div>
        {!collapsed && (
          <>
            <span className="font-semibold text-zinc-50 truncate">LinkClaw</span>
            <div className="ml-auto w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]" />
          </>
        )}
      </div>
      <nav className="flex-1 py-4 space-y-0.5 px-2">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-all duration-200 group",
                active
                  ? "bg-gradient-to-r from-blue-500/15 to-indigo-500/15 text-blue-400 border border-blue-500/20 shadow-md shadow-blue-500/10"
                  : "text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800/50 hover:pl-4"
              )}
            >
              <Icon className={cn(
                "w-4 h-4 flex-shrink-0 transition-colors",
                active ? "text-blue-400" : "group-hover:text-zinc-300"
              )} />
              {!collapsed && <span className="truncate">{label}</span>}
            </Link>
          );
        })}
      </nav>
      <div className="p-2 border-t border-zinc-800/50">
        <div className={cn(
          "px-3 py-2 rounded-md text-xs text-zinc-500 bg-zinc-800/30",
          collapsed && "text-center"
        )}>
          {!collapsed ? (
            <>
              <span className="block text-zinc-400 font-medium">v1.0.0</span>
              <span className="text-zinc-600">Build 2026</span>
            </>
          ) : (
            <span>⌘</span>
          )}
        </div>
      </div>
    </aside>
  );
}

function ThemeToggle() {
  const t = useTranslations();
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 transition-colors"
      title={mounted ? (theme === "dark" ?
        t("settings.themeLight", { defaultMessage: "Switch to light theme" }) :
        t("settings.themeDark", { defaultMessage: "Switch to dark theme" })) : undefined}
    >
      {mounted && theme === "dark" ? (
        <Sun className="w-4 h-4" />
      ) : (
        <Moon className="w-4 h-4" />
      )}
    </button>
  );
}

function LangToggle() {
  const t = useTranslations();
  const [locale, setLocale] = useState("zh");
  useEffect(() => {
    setLocale(localStorage.getItem("lc_locale") || "zh");
  }, []);

  const toggle = () => {
    const next = locale === "zh" ? "en" : "zh";
    localStorage.setItem("lc_locale", next);
    window.location.reload();
  };

  return (
    <button
      onClick={toggle}
      className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 transition-colors flex items-center gap-1"
      title={t("settings.language", { defaultMessage: "Switch language" })}
    >
      <Globe className="w-4 h-4" />
      <span className="text-xs font-mono">{locale === "zh" ? "中" : "EN"}</span>
    </button>
  );
}

function LogoutButton() {
  const t = useTranslations();
  const router = useRouter();

  const doLogout = async () => {
    try {
      const token = localStorage.getItem("lc_token");
      if (token) {
        await fetch("/api/v1/auth/logout", {
          method: "POST",
          headers: { Authorization: `Bearer ${token}` },
        });
      }
    } catch {
      // 即使接口失败也清除本地状态
    }
    destroyWSClient();
    localStorage.removeItem("lc_token");
    localStorage.removeItem("lc_agent_id");
    router.replace("/login");
  };

  const handleLogout = () => {
    toast(t("auth.logoutConfirm", { defaultMessage: "Confirm logout?" }), {
      action: {
        label: t("common.confirm", { defaultMessage: "Confirm" }),
        onClick: doLogout
      },
      cancel: {
        label: t("common.cancel", { defaultMessage: "Cancel" }),
        onClick: () => {}
      },
    });
  };

  return (
    <button
      onClick={handleLogout}
      className="p-1.5 rounded-md text-zinc-400 hover:text-red-400 hover:bg-zinc-800 transition-colors"
      title={t("auth.logout", { defaultMessage: "Logout" })}
    >
      <LogOut className="w-4 h-4" />
    </button>
  );
}

function Header({
  collapsed,
  onToggle,
}: {
  collapsed: boolean;
  onToggle: () => void;
}) {
  return (
    <header
      className={cn(
        "fixed top-0 right-0 h-14 bg-zinc-900/95 backdrop-blur-md border-b border-zinc-800/50 flex items-center px-4 z-30 transition-all duration-300 ease-in-out",
        collapsed ? "left-16" : "left-60"
      )}
    >
      <button
        onClick={onToggle}
        className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800/50 transition-all active:scale-95"
      >
        {collapsed ? <Menu className="w-4 h-4" /> : <X className="w-4 h-4" />}
      </button>
      <div className="flex-1" />
      <div className="flex items-center gap-1">
        <LangToggle />
        <ThemeToggle />
        <LogoutButton />
      </div>
    </header>
  );
}

export function Shell({
  children,
  noPadding = false,
}: {
  children: ReactNode;
  noPadding?: boolean;
}) {
  const [collapsed, setCollapsed] = useState(false);
  return (
    <div className="min-h-screen bg-zinc-950">
      <Sidebar collapsed={collapsed} />
      <Header collapsed={collapsed} onToggle={() => setCollapsed((v) => !v)} />
      <main
        className={cn(
          "pt-14 transition-all duration-300 ease-in-out",
          collapsed ? "pl-16" : "pl-60"
        )}
      >
        {noPadding ? children : <div className="p-6 space-y-6">{children}</div>}
      </main>
    </div>
  );
}
