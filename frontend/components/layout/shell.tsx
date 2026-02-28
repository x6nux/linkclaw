"use client";

import { ReactNode, useEffect, useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useTheme } from "next-themes";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { destroyWSClient } from "@/lib/ws-singleton";
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
} from "lucide-react";
import { useRouter } from "next/navigation";

const nav = [
  { href: "/dashboard", label: "概览", icon: LayoutDashboard },
  { href: "/agents", label: "Agent", icon: Bot },
  { href: "/prompts", label: "提示词", icon: ScrollText },
  { href: "/tasks", label: "任务", icon: CheckSquare },
  { href: "/messages", label: "消息", icon: MessageSquare },
  { href: "/knowledge", label: "知识库", icon: BookOpen },
  { href: "/memories", label: "记忆", icon: Brain },
  { href: "/llm", label: "LLM 网关", icon: Cpu },
  { href: "/organization", label: "组织架构", icon: Building2 },
  { href: "/observability", label: "可观测性", icon: Activity },
  { href: "/settings", label: "设置", icon: Settings },
];

function Sidebar({ collapsed }: { collapsed: boolean }) {
  const pathname = usePathname();
  return (
    <aside
      className={cn(
        "fixed left-0 top-0 h-full bg-zinc-900 border-r border-zinc-800 flex flex-col transition-all duration-200 z-40",
        collapsed ? "w-16" : "w-60"
      )}
    >
      <div className="flex items-center gap-2 px-4 h-14 border-b border-zinc-800">
        <Zap className="w-5 h-5 text-blue-500 flex-shrink-0" />
        {!collapsed && (
          <span className="font-semibold text-zinc-50 truncate">LinkClaw</span>
        )}
      </div>
      <nav className="flex-1 py-4 space-y-1 px-2">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors",
                active
                  ? "bg-blue-500/10 text-blue-400"
                  : "text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800"
              )}
            >
              <Icon className="w-4 h-4 flex-shrink-0" />
              {!collapsed && <span>{label}</span>}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}

function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);
  useEffect(() => setMounted(true), []);

  return (
    <button
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
      className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 transition-colors"
      title={mounted ? (theme === "dark" ? "切换为浅色" : "切换为深色") : undefined}
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
      title="切换语言"
    >
      <Globe className="w-4 h-4" />
      <span className="text-xs font-mono">{locale === "zh" ? "中" : "EN"}</span>
    </button>
  );
}

function LogoutButton() {
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
    toast("确认退出登录？", {
      action: { label: "退出", onClick: doLogout },
      cancel: { label: "取消", onClick: () => {} },
    });
  };

  return (
    <button
      onClick={handleLogout}
      className="p-1.5 rounded-md text-zinc-400 hover:text-red-400 hover:bg-zinc-800 transition-colors"
      title="退出登录"
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
        "fixed top-0 right-0 h-14 bg-zinc-900 border-b border-zinc-800 flex items-center px-4 z-30 transition-all duration-200",
        collapsed ? "left-16" : "left-60"
      )}
    >
      <button
        onClick={onToggle}
        className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-50 hover:bg-zinc-800 transition-colors"
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
          "pt-14 transition-all duration-200",
          collapsed ? "pl-16" : "pl-60"
        )}
      >
        {noPadding ? children : <div className="p-6">{children}</div>}
      </main>
    </div>
  );
}
