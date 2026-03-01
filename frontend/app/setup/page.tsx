"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";

const SLUG_REGEX = /^[a-z0-9-]+$/;

export default function SetupPage() {
  const t = useTranslations();
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
      setError(t("setup.passwordLength", { defaultMessage: "At least 8 characters required" }));
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
        setError(data.error || t("setup.error", { defaultMessage: "Initialization failed, please retry" }));
        return;
      }

      const data = await res.json();
      localStorage.setItem("lc_token", data.token);
      if (data.agent?.id) {
        localStorage.setItem("lc_agent_id", data.agent.id);
      }
      router.replace("/dashboard");
    } catch {
      setError(t("setup.networkError", { defaultMessage: "Network error, please try again" }));
    } finally {
      setLoading(false);
    }
  };

  const passwordLength = form.password.length;
  const passwordValid = passwordLength >= 8;
  const slugValid = form.companySlug.length > 0 && SLUG_REGEX.test(form.companySlug);

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
      <div className="w-full max-w-md p-8 bg-zinc-900 border border-zinc-800 rounded-xl">
        <h1 className="text-xl font-semibold text-zinc-50 mb-1">
          {t("setup.title", { defaultMessage: "Initialize LinkClaw" })}
        </h1>
        <p className="text-zinc-400 text-sm mb-6">
          {t("setup.subtitle", { defaultMessage: "First-time setup required" })}
        </p>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">
              {t("setup.companyName", { defaultMessage: "Company Name" })}
            </label>
            <input
              name="companyName"
              value={form.companyName}
              onChange={handleChange}
              autoComplete="organization"
              autoFocus
              placeholder={t("setup.companyNamePlaceholder", { defaultMessage: "e.g. My Virtual Company" })}
              required
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">
              {t("setup.companySlug", { defaultMessage: "Company Slug" })}
            </label>
            <input
              name="companySlug"
              value={form.companySlug}
              onChange={handleChange}
              placeholder={t("setup.companySlugPlaceholder", { defaultMessage: "e.g. my-company (lowercase + hyphens)" })}
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
              {t("setup.slugHint", { defaultMessage: "Only lowercase letters, numbers, and hyphens" })}
              {form.companySlug.length > 0 && (slugValid ? " ✓" : " ✗")}
            </p>
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">
              {t("setup.adminName", { defaultMessage: "Admin Username" })}
            </label>
            <input
              name="adminName"
              value={form.adminName}
              onChange={handleChange}
              autoComplete="username"
              placeholder={t("setup.adminNamePlaceholder", { defaultMessage: "e.g. Admin" })}
              required
              className="w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm text-zinc-400">
              {t("setup.password", { defaultMessage: "Admin Password" })}
            </label>
            <input
              type="password"
              name="password"
              value={form.password}
              onChange={handleChange}
              autoComplete="new-password"
              placeholder={t("setup.passwordPlaceholder", { defaultMessage: "At least 8 characters" })}
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
              {passwordLength}/8 {t("setup.characters", { defaultMessage: "characters" })}
              {passwordValid && " ✓"}
            </p>
          </div>
          {error && <p className="text-red-400 text-sm">{error}</p>}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors"
          >
            {loading
              ? t("setup.initializing", { defaultMessage: "Initializing..." })
              : t("setup.submit", { defaultMessage: "Initialize" })}
          </button>
        </form>
      </div>
    </div>
  );
}
