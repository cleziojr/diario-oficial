import { useCallback, useEffect, useState } from 'react'
import {
  createDocument,
  deleteDocument,
  getDocument,
  listDocuments,
  type Document,
} from './api/documents'
import './DocumentsPlayground.css'

const LIMIT_OPTIONS = [10, 20, 50, 100] as const

export function DocumentsPlayground() {
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [items, setItems] = useState<Document[]>([])
  const [listLoading, setListLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [filename, setFilename] = useState('')
  const [createBusy, setCreateBusy] = useState(false)

  const [selected, setSelected] = useState<Document | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    setListLoading(true)
    setError(null)
    try {
      const data = await listDocuments(page, limit)
      setItems(data.items)
    } catch (e) {
      setItems([])
      setError(e instanceof Error ? e.message : 'Falha ao listar documentos')
    } finally {
      setListLoading(false)
    }
  }, [page, limit])

  useEffect(() => {
    void refresh()
  }, [refresh])

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    const name = filename.trim()
    if (!name) {
      setError('Informe um nome de arquivo.')
      return
    }
    setCreateBusy(true)
    setError(null)
    try {
      await createDocument(name)
      setFilename('')
      await refresh()
      setSelected(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao criar')
    } finally {
      setCreateBusy(false)
    }
  }

  async function handleShowDetail(id: string) {
    setDetailError(null)
    setDetailLoading(true)
    setSelected(null)
    try {
      const doc = await getDocument(id)
      setSelected(doc)
    } catch (err) {
      setDetailError(err instanceof Error ? err.message : 'Falha ao buscar')
    } finally {
      setDetailLoading(false)
    }
  }

  async function handleDelete(id: string, label: string) {
    if (!window.confirm(`Excluir “${label}”?`)) {
      return
    }
    setError(null)
    try {
      await deleteDocument(id)
      if (selected?.id === id) {
        setSelected(null)
      }
      await refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao excluir')
    }
  }

  const canGoNext = items.length === limit
  const canGoPrev = page > 1

  return (
    <div className="documents-playground">
      <header className="documents-playground__header">
        <h1>Documentos (API)</h1>
        <p className="documents-playground__lede">
          Teste o CRUD de metadados contra <code>/api/v1/documents</code>. Em dev, o
          Vite encaminha para o backend em <code>:8080</code>.
        </p>
      </header>

      {error ? (
        <div className="documents-playground__banner documents-playground__banner--error" role="alert">
          {error}
        </div>
      ) : null}

      <section className="documents-playground__card">
        <h2>Novo documento</h2>
        <form className="documents-playground__form" onSubmit={handleCreate}>
          <label className="documents-playground__label">
            <span>Filename</span>
            <input
              type="text"
              name="filename"
              value={filename}
              onChange={(ev) => setFilename(ev.target.value)}
              placeholder="ex.: edital-2026-001.pdf"
              autoComplete="off"
            />
          </label>
          <button type="submit" className="documents-playground__btn documents-playground__btn--primary" disabled={createBusy}>
            {createBusy ? 'Salvando…' : 'Criar'}
          </button>
        </form>
      </section>

      <section className="documents-playground__card">
        <div className="documents-playground__toolbar">
          <h2>Lista</h2>
          <div className="documents-playground__pager">
            <label>
              <span className="visually-hidden">Itens por página</span>
              <select
                value={limit}
                onChange={(ev) => {
                  setLimit(Number(ev.target.value))
                  setPage(1)
                }}
              >
                {LIMIT_OPTIONS.map((n) => (
                  <option key={n} value={n}>
                    {n} / página
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              className="documents-playground__btn"
              disabled={!canGoPrev || listLoading}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Anterior
            </button>
            <span className="documents-playground__page-indicator">
              Página {page}
            </span>
            <button
              type="button"
              className="documents-playground__btn"
              disabled={!canGoNext || listLoading}
              onClick={() => setPage((p) => p + 1)}
            >
              Próxima
            </button>
            <button
              type="button"
              className="documents-playground__btn"
              disabled={listLoading}
              onClick={() => void refresh()}
            >
              Atualizar
            </button>
          </div>
        </div>

        {listLoading ? (
          <p className="documents-playground__muted">Carregando…</p>
        ) : items.length === 0 ? (
          <p className="documents-playground__muted">Nenhum registro nesta página.</p>
        ) : (
          <div className="documents-playground__table-wrap">
            <table className="documents-playground__table">
              <thead>
                <tr>
                  <th>Filename</th>
                  <th>Criado em</th>
                  <th>ID</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items.map((row) => (
                  <tr key={row.id}>
                    <td>{row.filename}</td>
                    <td className="documents-playground__mono">{formatDate(row.created_at)}</td>
                    <td className="documents-playground__mono documents-playground__id-cell" title={row.id}>
                      {shortId(row.id)}
                    </td>
                    <td className="documents-playground__actions">
                      <button
                        type="button"
                        className="documents-playground__btn documents-playground__btn--small"
                        onClick={() => void handleShowDetail(row.id)}
                      >
                        Ver (GET)
                      </button>
                      <button
                        type="button"
                        className="documents-playground__btn documents-playground__btn--small documents-playground__btn--danger"
                        onClick={() => void handleDelete(row.id, row.filename)}
                      >
                        Excluir
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="documents-playground__card">
        <h2>Detalhe (GET por id)</h2>
        {detailLoading ? (
          <p className="documents-playground__muted">Carregando…</p>
        ) : detailError ? (
          <p className="documents-playground__banner documents-playground__banner--error">{detailError}</p>
        ) : selected ? (
          <dl className="documents-playground__detail">
            <dt>ID</dt>
            <dd className="documents-playground__mono">{selected.id}</dd>
            <dt>Filename</dt>
            <dd>{selected.filename}</dd>
            <dt>created_at</dt>
            <dd className="documents-playground__mono">{selected.created_at}</dd>
          </dl>
        ) : (
          <p className="documents-playground__muted">Clique em “Ver (GET)” na lista.</p>
        )}
      </section>
    </div>
  )
}

function shortId(uuid: string): string {
  if (uuid.length <= 13) {
    return uuid
  }
  return `${uuid.slice(0, 8)}…${uuid.slice(-4)}`
}

function formatDate(iso: string): string {
  if (!iso) {
    return '—'
  }
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) {
    return iso
  }
  return d.toLocaleString()
}
