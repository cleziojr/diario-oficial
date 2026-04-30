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

function apiPrefix(): string {
  const base = import.meta.env.VITE_API_BASE as string | undefined
  return (base ?? '').replace(/\/$/, '')
}

async function readErrorMessage(res: Response): Promise<string> {
  try {
    const j: unknown = await res.json()
    if (
      j &&
      typeof j === 'object' &&
      'error' in j &&
      typeof (j as { error: unknown }).error === 'string'
    ) {
      return (j as { error: string }).error
    }
  } catch {
    /* ignore */
  }
  return res.statusText || `HTTP ${res.status}`
}

export async function listDocuments(
  page: number,
  limit: number,
): Promise<ListDocumentsResponse> {
  const q = new URLSearchParams({
    page: String(page),
    limit: String(limit),
  })
  const res = await fetch(`${apiPrefix()}/api/v1/documents?${q}`)
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  return res.json() as Promise<ListDocumentsResponse>
}

export async function getDocument(id: string): Promise<Document> {
  const res = await fetch(`${apiPrefix()}/api/v1/documents/${encodeURIComponent(id)}`)
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  return res.json() as Promise<Document>
}

export async function createDocument(filename: string): Promise<Document> {
  const res = await fetch(`${apiPrefix()}/api/v1/documents`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filename }),
  })
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
  return res.json() as Promise<Document>
}

export async function deleteDocument(id: string): Promise<void> {
  const res = await fetch(`${apiPrefix()}/api/v1/documents/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
  if (res.status === 204) {
    return
  }
  if (!res.ok) {
    throw new Error(await readErrorMessage(res))
  }
}
