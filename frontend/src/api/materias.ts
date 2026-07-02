import { apiFetch, apiPrefix, authHeader } from './client'

export type Entities = {
  people: string[]
  orgs: string[]
}

export type DocumentRef = {
  id: string
  filename: string
  created_at: string
}

export type Materia = {
  id: string
  document_id: string
  title: string
  summary: string
  category: string
  tags: string[]
  entities: Entities
  monetary_values: string[]
  dates: string[]
  page: number
  document: DocumentRef
  created_at: string
}

export type SearchResponse = {
  items: Materia[]
  page: number
  limit: number
  total: number
}

export type TagCount = {
  tag: string
  label: string
  count: number
}

export type IngestResponse = {
  document_id: string
  materias_count: number
  pages: number
}

export type SearchParams = {
  q?: string
  tags?: string[]
  page?: number
  limit?: number
}

export async function searchMaterias(params: SearchParams): Promise<SearchResponse> {
  const q = new URLSearchParams()
  if (params.q) q.set('q', params.q)
  if (params.tags && params.tags.length > 0) q.set('tags', params.tags.join(','))
  q.set('page', String(params.page ?? 1))
  q.set('limit', String(params.limit ?? 20))
  const res = await apiFetch(`/api/v1/search?${q}`)
  return res.json() as Promise<SearchResponse>
}

export async function listTags(): Promise<TagCount[]> {
  const res = await apiFetch('/api/v1/tags')
  return res.json() as Promise<TagCount[]>
}

export async function getMateria(id: string): Promise<Materia> {
  const res = await apiFetch(`/api/v1/materias/${encodeURIComponent(id)}`)
  return res.json() as Promise<Materia>
}

export async function listDocumentMaterias(documentId: string): Promise<Materia[]> {
  const res = await apiFetch(`/api/v1/documents/${encodeURIComponent(documentId)}/materias`)
  return res.json() as Promise<Materia[]>
}

export async function ingestDocument(
  file: File,
  filename?: string,
  onProgress?: (percent: number) => void,
  onProcessing?: () => void,
): Promise<IngestResponse> {
  const form = new FormData()
  form.append('file', file)
  if (filename) form.append('filename', filename)

  if (!onProgress) {
    onProcessing?.()
    const res = await apiFetch('/api/v1/ingest', { method: 'POST', body: form })
    return res.json() as Promise<IngestResponse>
  }

  return new Promise<IngestResponse>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', `${apiPrefix()}/api/v1/ingest`)
    for (const [k, v] of Object.entries(authHeader())) {
      xhr.setRequestHeader(k, v)
    }
    xhr.upload.onprogress = (e) => {
      if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100))
    }
    // Bytes enviados: a partir daqui o servidor extrai o PDF e a IA categoriza.
    xhr.upload.onload = () => onProcessing?.()
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText) as IngestResponse)
        } catch {
          reject(new Error('Resposta inválida do servidor'))
        }
      } else {
        let message = xhr.statusText || `HTTP ${xhr.status}`
        try {
          const j = JSON.parse(xhr.responseText) as { error?: string }
          if (j.error) message = j.error
        } catch {
          /* ignore */
        }
        reject(new Error(message))
      }
    }
    xhr.onerror = () => reject(new Error('Falha de rede ao enviar o arquivo'))
    xhr.send(form)
  })
}
