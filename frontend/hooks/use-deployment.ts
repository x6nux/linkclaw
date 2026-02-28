"use client";

import { useState } from "react";
import type { AgentDeployment, DeployRequest } from "@/lib/types";

export function useDeployment() {
  const [isDeploying, setIsDeploying] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function deploy(agentId: string, req: DeployRequest): Promise<AgentDeployment | null> {
    setIsDeploying(true);
    setError(null);
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/agents/${agentId}/deploy`, {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify(req),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "部署失败");
      return data as AgentDeployment;
    } catch (e) {
      setError(e instanceof Error ? e.message : "部署失败");
      return null;
    } finally {
      setIsDeploying(false);
    }
  }

  async function getDeployment(agentId: string): Promise<AgentDeployment | null> {
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/agents/${agentId}/deployment`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.status === 404) return null;
      const data = await res.json();
      return data as AgentDeployment;
    } catch {
      return null;
    }
  }

  async function stopDeployment(agentId: string): Promise<boolean> {
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/agents/${agentId}/deployment`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}` },
      });
      return res.ok;
    } catch {
      return false;
    }
  }

  async function rebuildDeployment(agentId: string): Promise<{ deployment: AgentDeployment; newApiKey: string } | null> {
    setIsDeploying(true);
    setError(null);
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch(`/api/v1/agents/${agentId}/deployment/rebuild`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "重建失败");
      return {
        deployment: data.deployment as AgentDeployment,
        newApiKey: (data.new_api_key ?? data.newApiKey) as string,
      };
    } catch (e) {
      setError(e instanceof Error ? e.message : "重建失败");
      return null;
    } finally {
      setIsDeploying(false);
    }
  }

  return { deploy, getDeployment, stopDeployment, rebuildDeployment, isDeploying, error };
}
