import { Link } from 'react-router-dom'
import type { Materia } from '../api/materias'
import { tagLabel } from '../api/taxonomy'
import './MateriaCard.css'

function formatDate(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleDateString('pt-BR')
}

export function MateriaCard({ materia }: { materia: Materia }) {
  return (
    <article className="materia-card">
      <div className="materia-card__header">
        <span className="materia-card__badge">{tagLabel(materia.category)}</span>
        <span className="materia-card__date">{formatDate(materia.created_at)}</span>
      </div>
      <h2>
        <Link to={`/materia/${materia.id}`}>{materia.title}</Link>
      </h2>
      <p className="materia-card__summary">{materia.summary}</p>
      <div className="materia-card__footer">
        <span>{materia.document.filename}</span>
        <span>pág. {materia.page}</span>
      </div>
      {materia.tags.length > 0 && (
        <ul className="materia-card__tags">
          {materia.tags.map((t) => (
            <li key={t}>{tagLabel(t)}</li>
          ))}
        </ul>
      )}
    </article>
  )
}
