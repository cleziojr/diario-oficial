import { apiFetch } from './client'

export type Document = {
  id: string
  filename: string
  created_at: string
}

export type ListDocumentsResponse = {
  items: Document[]
  page: number
  limit: number
}

export async function listDocuments(
  page: number,
  limit: number,
): Promise<ListDocumentsResponse> {
  const q = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  })
  const res = await apiFetch(`/api/v1/documents?${q}`)
  return res.json() as Promise<ListDocumentsResponse>
}

export async function getDocument(id: string): Promise<Document> {
  const res = await apiFetch(`/api/v1/documents/${encodeURIComponent(id)}`)
  return res.json() as Promise<Document>
}

export async function createDocument(filename: string): Promise<Document> {
  const res = await apiFetch('/api/v1/documents', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filename }),
  })
  return res.json() as Promise<Document>
}

export async function deleteDocument(id: string): Promise<void> {
  await apiFetch(`/api/v1/documents/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
