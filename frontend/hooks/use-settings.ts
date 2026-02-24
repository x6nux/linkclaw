import useSWR from "swr";
import { api } from "@/lib/api";
import type { CompanySettings } from "@/lib/types";

export function useSettings() {
  const { data, error, isLoading, mutate } = useSWR<CompanySettings>(
    "/api/v1/settings",
    (url: string) => api.get<CompanySettings>(url)
  );
  return { settings: data ?? null, isLoading, error, mutate };
}

export async function updateSettings(s: CompanySettings) {
  return api.put<{ ok: boolean }>("/api/v1/settings", s);
}
