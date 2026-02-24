"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const SLUG_REGEX = /^[a-z0-9-]+$/;

export default function SetupPage() {
  const router = useRouter();
  const [form, setForm] = useState({
    companyName: "",
    companySlug: "",
    adminName: "",
    password: "",
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [e.target.name]: e.target.value }));
    if (error) setError("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (form.password.length < 8) {
      setError("密码至少需要 8 位");
      return;
    }
    setLoading(true);
    setError("");

    try {
      const res = await fetch("/api/v1/setup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(form),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        setError(data.error || "初始化失败，请重试");
        return;
      }

      const data = await res.json();
      localStorage.setItem("lc_token", data.token);
      if (data.agent?.id) {
        localStorage.setItem("lc_agent_id", data.agent.id);
      }
      router.replace("/dashboard");
    } catch {
      setError("网络错误，请重试");
    } finally {
      setLoading(false);
    }
  };

  const passwordLength = form.password.length;
  const passwordValid = passwordLength >= 8;
  const slugValid =
    form.companySlug.length > 0 && SLUG_REGEX.test(form.companySlug);

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
      <div className="w-full max-w-md p-8 bg-zinc-900 border border-zinc-800 rounded-xl">
        <h1 className="text-xl font-semibold text-zinc-50 mb-1">初始化 LinkClaw</h1>
        <p className="text-zinc-400 text-sm mb-6">首次使用，请完成初始化设置</p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">公司名称</label>
            <input
              name="companyName"
              value={form.companyName}
              onChange={handleChange}
              autoComplete="organization"
              autoFocus
              placeholder="例：我的虚拟公司"
              required
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">公司标识（Slug）</label>
            <input
              name="companySlug"
              value={form.companySlug}
              onChange={handleChange}
              placeholder="例：my-company（英文小写+短横线）"
              required
              pattern="[a-z0-9-]+"
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
            <p
              className={`text-xs mt-1 transition-colors ${
                form.companySlug.length === 0
                  ? "text-zinc-500"
                  : slugValid
                    ? "text-green-400"
                    : "text-red-400"
              }`}
            >
              仅允许小写字母、数字和短横线
              {form.companySlug.length > 0 && (slugValid ? " ✓" : " ✗")}
            </p>
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">管理员用户名</label>
            <input
              name="adminName"
              value={form.adminName}
              onChange={handleChange}
              autoComplete="username"
              placeholder="例：Admin"
              required
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">管理员密码</label>
            <input
              type="password"
              name="password"
              value={form.password}
              onChange={handleChange}
              autoComplete="new-password"
              placeholder="至少 8 位"
              required
              minLength={8}
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
            <p
              className={`text-xs mt-1 transition-colors ${
                passwordLength === 0
                  ? "text-zinc-500"
                  : passwordValid
                    ? "text-green-400"
                    : "text-zinc-400"
              }`}
            >
              {passwordLength}/8 字符{passwordValid && " ✓"}
            </p>
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors"
          >
            {loading ? "初始化中..." : "完成初始化"}
          </button>
        </form>
      </div>
    </div>
  );
}
