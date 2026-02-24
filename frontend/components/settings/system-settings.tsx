"use client";

import { useState, useEffect } from "react";
import { Server } from "lucide-react";
import { useSettings, updateSettings } from "@/hooks/use-settings";
import type { CompanySettings } from "@/lib/types";

const fields: {
  key: keyof CompanySettings;
  label: string;
  placeholder: string;
  hint: string;
}[] = [
  {
    key: "publicDomain",
    label: "公网域名",
    placeholder: "example.com",
    hint: "用于生成对外访问地址",
  },
  {
    key: "agentWsUrl",
    label: "Agent 连接地址",
    placeholder: "ws://example.com/api/v1/agents/me/ws",
    hint: "留空则根据公网域名自动生成",
  },
  {
    key: "mcpPublicUrl",
    label: "公网 MCP 地址",
    placeholder: "https://example.com/mcp/sse",
    hint: "留空则根据公网域名自动生成",
  },
  {
    key: "nanoclawImage",
    label: "NanoClaw 镜像名称",
    placeholder: "nanoclaw:latest",
    hint: "Docker 镜像全名",
  },
  {
    key: "openclawPluginUrl",
    label: "OpenClaw 插件地址",
    placeholder: "ghcr.io/qwibitai/openclaw:latest",
    hint: "Docker 镜像或下载地址",
  },
];

const emptySettings: CompanySettings = {
  publicDomain: "",
  agentWsUrl: "",
  mcpPublicUrl: "",
  nanoclawImage: "",
  openclawPluginUrl: "",
};

export function SystemSettings() {
  const { settings, isLoading, mutate } = useSettings();
  const [form, setForm] = useState<CompanySettings>(emptySettings);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    if (settings) setForm(settings);
  }, [settings]);

  const handleChange = (key: keyof CompanySettings, value: string) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    setError("");
    setSuccess("");
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setSaving(true);
    try {
      await updateSettings(form);
      await mutate(form, false);
      setSuccess("保存成功");
    } catch (err) {
      setError(err instanceof Error ? err.message : "保存失败");
    } finally {
      setSaving(false);
    }
  };

  const inputClass =
    "w-full px-3 py-2 bg-zinc-800 border border-zinc-700 rounded-md text-zinc-50 placeholder-zinc-500 text-sm focus:outline-none focus:border-blue-500 transition-colors";

  if (isLoading) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
        <div className="flex items-center gap-2">
          <Server className="w-4 h-4 text-zinc-400" />
          <h2 className="text-sm font-medium text-zinc-200">系统配置</h2>
        </div>
        <div className="space-y-3 animate-pulse">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i}>
              <div className="h-3 w-24 bg-zinc-800 rounded mb-2" />
              <div className="h-9 w-full bg-zinc-800 rounded" />
            </div>
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-6 space-y-4">
      <div className="flex items-center gap-2">
        <Server className="w-4 h-4 text-zinc-400" />
        <h2 className="text-sm font-medium text-zinc-200">系统配置</h2>
      </div>
      <form onSubmit={handleSubmit} className="space-y-3">
        {fields.map(({ key, label, placeholder, hint }) => (
          <div key={key}>
            <label className="text-xs text-zinc-500 mb-1 block">{label}</label>
            <input
              type="text"
              value={form[key]}
              onChange={(e) => handleChange(key, e.target.value)}
              placeholder={placeholder}
              className={inputClass}
            />
            <p className="text-xs text-zinc-600 mt-0.5">{hint}</p>
          </div>
        ))}
        {error && <p className="text-red-400 text-xs">{error}</p>}
        {success && <p className="text-green-400 text-xs">{success}</p>}
        <button
          type="submit"
          disabled={saving}
          className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white rounded-md text-sm font-medium transition-colors"
        >
          {saving ? "保存中..." : "保存配置"}
        </button>
      </form>
    </div>
  );
}
