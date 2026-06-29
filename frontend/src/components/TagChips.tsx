import type { TagCount } from '../api/materias'
import './TagChips.css'

type TagChipsProps = {
  tags: TagCount[]
  selected: string[]
  onToggle: (tag: string) => void
}

export function TagChips({ tags, selected, onToggle }: TagChipsProps) {
  if (tags.length === 0) return null

  return (
    <div className="tag-chips" role="group" aria-label="Filtrar por tag">
      {tags.map((t) => {
        const isSelected = selected.includes(t.tag)
        return (
          <button
            key={t.tag}
            type="button"
            className={`tag-chip${isSelected ? ' tag-chip--selected' : ''}`}
            aria-pressed={isSelected}
            onClick={() => onToggle(t.tag)}
          >
            {t.label} <span className="tag-chip__count">{t.count}</span>
          </button>
        )
      })}
    </div>
  )
}
