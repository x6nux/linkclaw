"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { Zap } from "lucide-react";
import { useTranslations } from "next-intl";

type Mode = "login" | "reset";

export default function LoginPage() {
  const t = useTranslations();
  const router = useRouter();
  const [mode, setMode] = useState<Mode>("login");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [resetSecret, setResetSecret] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [shakeKey, setShakeKey] = useState(0);
  const errorRef = useRef<HTMLDivElement>(null);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      const res = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, password }),
      });

      if (!res.ok) {
        setError(t("auth.error", { defaultMessage: "Invalid username or password" }));
        setShakeKey((k) => k + 1);
        return;
      }

      const data = await res.json();
      localStorage.setItem("lc_token", data.token);
      if (data.agent?.id) {
        localStorage.setItem("lc_agent_id", data.agent.id);
      }
      router.replace("/dashboard");
    } catch {
      setError(t("auth.networkError", { defaultMessage: "Network error, please retry" }));
      setShakeKey((k) => k + 1);
    } finally {
      setLoading(false);
    }
  };

  const handleReset = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (newPassword !== confirmPassword) {
      setError(t("auth.passwordMismatch", { defaultMessage: "Passwords do not match" }));
      setShakeKey((k) => k + 1);
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/v1/auth/reset-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, resetSecret, newPassword }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        if (res.status === 503) {
          setError(t("auth.resetKeyNotEnabled", { defaultMessage: "Password reset not enabled" }));
        } else {
          setError(data.error === "invalid credentials" 
            ? t("auth.invalidCredentials", { defaultMessage: "Invalid username or reset secret" })
            : t("auth.resetFailed", { defaultMessage: "Reset failed, please retry" }));
        }
        setShakeKey((k) => k + 1);
        return;
      }

      setSuccess(t("auth.resetSuccess", { defaultMessage: "Password reset successfully" }));
      setMode("login");
      setPassword("");
      setResetSecret("");
      setNewPassword("");
      setConfirmPassword("");
    } catch {
      setError(t("auth.networkError", { defaultMessage: "Network error, please retry" }));
      setShakeKey((k) => k + 1);
    } finally {
      setLoading(false);
    }
  };

  const switchMode = (target: Mode) => {
    setMode(target);
    setError("");
    setSuccess("");
  };

  useEffect(() => {
    fetch("/api/v1/setup/status")
      .then((r) => r.json())
      .then((data) => {
        if (!data.initialized) router.replace("/setup");
      })
      .catch(() => {});
  }, [router]);

  useEffect(() => {
    if (!errorRef.current) return;
    const el = errorRef.current;
    const handler = () => el.classList.remove("animate-shake");
    el.addEventListener("animationend", handler);
    return () => el.removeEventListener("animationend", handler);
  }, [shakeKey]);

  const inputBorderClass = error ? "border-red-500/50" : "border-zinc-700";
  const inputClass = `w-full px-3 py-2 bg-zinc-800 border ${inputBorderClass} rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors`;

  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center">
      <div className="w-full max-w-sm p-8 bg-zinc-900 border border-zinc-800 rounded-xl">
        <div className="flex items-center gap-2 mb-6">
          <Zap className="w-6 h-6 text-blue-500" />
          <h1 className="text-xl font-semibold text-zinc-50">
            {mode === "login" 
              ? t("auth.title", { defaultMessage: "Login to LinkClaw" })
              : t("auth.resetTitle", { defaultMessage: "Reset Password" })}
          </h1>
        </div>

        {success && (
          <div className="mb-4 p-3 bg-green-500/10 border border-green-500/20 rounded-md">
            <p className="text-green-400 text-sm">{success}</p>
          </div>
        )}

        {mode === "login" ? (
          <form onSubmit={handleLogin} className="space-y-4">
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.name", { defaultMessage: "Username" })}
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => { setName(e.target.value); if (error) setError(""); }}
                autoComplete="username"
                autoFocus
                placeholder={t("auth.namePlaceholder", { defaultMessage: "Enter your username" })}
                required
                className={inputClass}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.password", { defaultMessage: "Password" })}
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => { setPassword(e.target.value); if (error) setError(""); }}
                autoComplete="current-password"
                placeholder={t("auth.passwordPlaceholder", { defaultMessage: "Enter password" })}
                required
                className={inputClass}
              />
            </div>
            {error && (
              <div key={shakeKey} ref={errorRef} className="animate-shake">
                <p className="text-red-400 text-sm">{error}</p>
              </div>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors"
            >
              {loading 
                ? t("auth.loggingIn", { defaultMessage: "Logging in..." })
                : t("auth.submit", { defaultMessage: "Login" })}
            </button>
            <button
              type="button"
              onClick={() => switchMode("reset")}
              className="w-full text-center text-xs text-zinc-500 hover:text-zinc-300 transition-colors"
            >
              {t("auth.forgotPassword", { defaultMessage: "Forgot password?" })}
            </button>
          </form>
        ) : (
          <form onSubmit={handleReset} className="space-y-4">
            <p className="text-xs text-zinc-500">
              {t("auth.resetSecretHint", { defaultMessage: "Find your reset secret in .env" })}
            </p>
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.name", { defaultMessage: "Username" })}
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => { setName(e.target.value); if (error) setError(""); }}
                autoComplete="username"
                autoFocus
                placeholder={t("auth.namePlaceholder", { defaultMessage: "Enter your username" })}
                required
                className={inputClass}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.resetSecret", { defaultMessage: "Reset Secret" })}
              </label>
              <input
                type="password"
                value={resetSecret}
                onChange={(e) => { setResetSecret(e.target.value); if (error) setError(""); }}
                autoComplete="off"
                placeholder={t("auth.resetSecretPlaceholder", { defaultMessage: "Enter reset secret" })}
                required
                className={inputClass}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.newPassword", { defaultMessage: "New Password" })}
              </label>
              <input
                type="password"
                value={newPassword}
                onChange={(e) => { setNewPassword(e.target.value); if (error) setError(""); }}
                autoComplete="new-password"
                placeholder={t("auth.newPasswordPlaceholder", { defaultMessage: "At least 8 characters" })}
                required
                minLength={8}
                className={inputClass}
              />
            </div>
            <div className="space-y-1">
              <label className="text-sm text-zinc-400">
                {t("auth.confirmNewPassword", { defaultMessage: "Confirm New Password" })}
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => { setConfirmPassword(e.target.value); if (error) setError(""); }}
                autoComplete="new-password"
                placeholder={t("auth.confirmNewPasswordPlaceholder", { defaultMessage: "Enter again" })}
                required
                minLength={8}
                className={inputClass}
              />
            </div>
            {error && (
              <div key={shakeKey} ref={errorRef} className="animate-shake">
                <p className="text-red-400 text-sm">{error}</p>
              </div>
            )}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors"
            >
              {loading 
                ? t("auth.resetting", { defaultMessage: "Resetting..." })
                : t("auth.resetPassword", { defaultMessage: "Reset Password" })}
            </button>
            <button
              type="button"
              onClick={() => switchMode("login")}
              className="w-full text-center text-xs text-zinc-500 hover:text-zinc-300 transition-colors"
            >
              {t("auth.backToLogin", { defaultMessage: "Back to login" })}
            </button>
          </form>
        )}
      </div>
    </div>
  );
}
