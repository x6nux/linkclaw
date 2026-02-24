import useSWR, { mutate } from "swr";
import { api } from "@/lib/api";

export interface KnowledgeDoc {
  id: string;
  companyId: string;
  title: string;
  content: string;
  tags: string[];
  authorId?: string;
  createdAt: string;
  updatedAt: string;
}

const fetcher = (url: string) =>
  api.get<{ docs: KnowledgeDoc[]; total: number }>(url).then((r) => r);

export function useKnowledgeDocs(search?: string) {
  const url = search
    ? `/api/v1/knowledge/search?q=${encodeURIComponent(search)}&limit=50`
    : `/api/v1/knowledge?limit=50`;
  const { data, error, isLoading } = useSWR(url, fetcher);
  return {
    docs: search ? (data as unknown as KnowledgeDoc[]) : data?.docs ?? [],
    total: data?.total ?? 0,
    isLoading,
    error,
    refresh: () => mutate(url),
  };
}

export function useKnowledgeDoc(id: string | null) {
  const { data, error, isLoading } = useSWR(
    id ? `/api/v1/knowledge/${id}` : null,
    (url: string) => api.get<KnowledgeDoc>(url)
  );
  return { doc: data, isLoading, error };
}

export async function createDoc(title: string, content: string, tags: string[]) {
  return api.post<KnowledgeDoc>("/api/v1/knowledge", { title, content, tags });
}

export async function updateDoc(id: string, title: string, content: string, tags: string[]) {
  return api.put<KnowledgeDoc>(`/api/v1/knowledge/${id}`, { title, content, tags });
}

export async function deleteDoc(id: string) {
  return api.delete(`/api/v1/knowledge/${id}`);
}
