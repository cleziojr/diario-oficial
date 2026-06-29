import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { searchMaterias, listTags, listDocumentMaterias } from '../api/materias'
import type { Materia, TagCount } from '../api/materias'
import { TagChips } from '../components/TagChips'
import { MateriaCard } from '../components/MateriaCard'
import { Pagination } from '../components/Pagination'
import './SearchPage.css'

const LIMIT = 20

export function SearchPage() {
  const [params, setParams] = useSearchParams()
  const q = params.get('q') ?? ''
  const selectedTags = params.get('tags')?.split(',').filter(Boolean) ?? []
  const page = Number(params.get('page') ?? '1') || 1
  const documentId = params.get('document_id')

  const [qInput, setQInput] = useState(q)
  const [tags, setTags] = useState<TagCount[]>([])
  const [items, setItems] = useState<Materia[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setQInput(q)
  }, [q])

  useEffect(() => {
    listTags()
      .then(setTags)
      .catch(() => setTags([]))
  }, [])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    setError(null)

    const request = documentId
      ? listDocumentMaterias(documentId).then((materias) => ({
          items: materias,
          total: materias.length,
        }))
      : searchMaterias({ q, tags: selectedTags, page, limit: LIMIT })

    request
      .then((res) => {
        if (cancelled) return
        setItems(res.items)
        setTotal(res.total)
      })
      .catch((err: Error) => {
        if (cancelled) return
        setError(err.message)
        setItems([])
        setTotal(0)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [q, selectedTags.join(','), page, documentId])

  function updateParams(next: {
    q?: string
    tags?: string[]
    page?: number
    documentId?: string | null
  }) {
    const merged = new URLSearchParams(params)
    if (next.q !== undefined) {
      if (next.q) merged.set('q', next.q)
      else merged.delete('q')
    }
    if (next.tags !== undefined) {
      if (next.tags.length > 0) merged.set('tags', next.tags.join(','))
      else merged.delete('tags')
    }
    if (next.documentId !== undefined) {
      if (next.documentId) merged.set('document_id', next.documentId)
      else merged.delete('document_id')
    }
    merged.set('page', String(next.page ?? 1))
    setParams(merged)
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    updateParams({ q: qInput, tags: selectedTags, page: 1 })
  }

  function handleToggleTag(tag: string) {
    const next = selectedTags.includes(tag)
      ? selectedTags.filter((t) => t !== tag)
      : [...selectedTags, tag]
    updateParams({ q, tags: next, page: 1 })
  }

  return (
    <div className="search-page">
      <h1>Busca de matérias</h1>

      {documentId && (
        <p className="search-page__filter-banner">
          Mostrando matérias do documento enviado.{' '}
          <button type="button" onClick={() => updateParams({ q: '', tags: [], page: 1, documentId: null })}>
            Limpar filtro
          </button>
        </p>
      )}

      <form className="search-page__form" onSubmit={handleSubmit}>
        <input
          type="search"
          placeholder="Buscar por palavra-chave (ex.: licitação)"
          value={qInput}
          onChange={(e) => setQInput(e.target.value)}
          aria-label="Buscar matérias"
        />
        <button type="submit">Buscar</button>
      </form>

      <TagChips tags={tags} selected={selectedTags} onToggle={handleToggleTag} />

      {loading && <p className="search-page__status">Carregando...</p>}
      {!loading && error && <p className="search-page__status search-page__status--error">{error}</p>}
      {!loading && !error && items.length === 0 && (
        <p className="search-page__status">Nenhuma matéria encontrada.</p>
      )}

      {!loading && !error && items.length > 0 && (
        <>
          <ul className="search-page__results">
            {items.map((m) => (
              <li key={m.id}>
                <MateriaCard materia={m} />
              </li>
            ))}
          </ul>
          {!documentId && (
            <Pagination
              page={page}
              limit={LIMIT}
              total={total}
              onPageChange={(p) => updateParams({ q, tags: selectedTags, page: p })}
            />
          )}
        </>
      )}
    </div>
  )
}
