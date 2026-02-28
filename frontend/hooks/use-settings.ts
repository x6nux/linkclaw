import useSWR from "swr";
import { api } from "@/lib/api";
import type { CompanySettings } from "@/lib/types";

export interface CompanySettingsPayload {
  public_domain: string;
  agent_ws_url: string;
  mcp_public_url: string;
  nanoclaw_image: string;
  openclaw_plugin_url: string;
  embedding_base_url: string;
  embedding_model: string;
  embedding_api_key: string;
}

export function useSettings() {
  const { data, error, isLoading, mutate } = useSWR<CompanySettings>(
    "/api/v1/settings",
    (url: string) => api.get<CompanySettings>(url)
  );
  return { settings: data ?? null, isLoading, error, mutate };
}

export async function updateSettings(s: CompanySettingsPayload) {
  return api.put<{ ok: boolean }>("/api/v1/settings", s);
}
