import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getMateria } from '../api/materias'
import type { Materia } from '../api/materias'
import { tagLabel } from '../api/taxonomy'
import './MateriaDetailPage.css'

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString('pt-BR')
}

export function MateriaDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [materia, setMateria] = useState<Materia | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    let cancelled = false
    setLoading(true)
    setError(null)
    getMateria(id)
      .then((m) => {
        if (!cancelled) setMateria(m)
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id])

  if (loading) return <p className="materia-detail__status">Carregando...</p>
  if (error) return <p className="materia-detail__status materia-detail__status--error">{error}</p>
  if (!materia) return null

  return (
    <article className="materia-detail">
      <Link to="/" className="materia-detail__back">
        ← Voltar à busca
      </Link>
      <span className="materia-detail__badge">{tagLabel(materia.category)}</span>
      <h1>{materia.title}</h1>
      <p className="materia-detail__summary">{materia.summary}</p>

      {materia.tags.length > 0 && (
        <ul className="materia-detail__tags">
          {materia.tags.map((t) => (
            <li key={t}>{tagLabel(t)}</li>
          ))}
        </ul>
      )}

      <dl className="materia-detail__meta">
        <div>
          <dt>Documento</dt>
          <dd>{materia.document.filename}</dd>
        </div>
        <div>
          <dt>Página</dt>
          <dd>{materia.page}</dd>
        </div>
        <div>
          <dt>Publicado em</dt>
          <dd>{formatDate(materia.document.created_at)}</dd>
        </div>
        <div>
          <dt>Indexado em</dt>
          <dd>{formatDate(materia.created_at)}</dd>
        </div>
      </dl>

      {(materia.entities.people.length > 0 || materia.entities.orgs.length > 0) && (
        <section>
          <h2>Entidades</h2>
          {materia.entities.people.length > 0 && (
            <p>
              <strong>Pessoas:</strong> {materia.entities.people.join(', ')}
            </p>
          )}
          {materia.entities.orgs.length > 0 && (
            <p>
              <strong>Órgãos:</strong> {materia.entities.orgs.join(', ')}
            </p>
          )}
        </section>
      )}

      {materia.monetary_values.length > 0 && (
        <section>
          <h2>Valores</h2>
          <ul>
            {materia.monetary_values.map((v, i) => (
              <li key={i}>{v}</li>
            ))}
          </ul>
        </section>
      )}

      {materia.dates.length > 0 && (
        <section>
          <h2>Datas citadas</h2>
          <ul>
            {materia.dates.map((d, i) => (
              <li key={i}>{d}</li>
            ))}
          </ul>
        </section>
      )}

      <p className="materia-detail__source">
        Origem: {materia.document.filename} (pág. {materia.page})
      </p>
    </article>
  )
}
