-- Matérias categorizadas pela LLM, indexadas para busca pública por tag + texto.
CREATE TABLE IF NOT EXISTS materias (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id     UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    title           TEXT        NOT NULL DEFAULT '',
    summary         TEXT        NOT NULL DEFAULT '',
    body            TEXT        NOT NULL DEFAULT '',
    category        TEXT        NOT NULL DEFAULT 'outros',
    tags            TEXT[]      NOT NULL DEFAULT '{}',
    entities        JSONB       NOT NULL DEFAULT '{}',
    monetary_values JSONB       NOT NULL DEFAULT '[]',
    dates           JSONB       NOT NULL DEFAULT '[]',
    page            INT,
    search_vector   TSVECTOR    GENERATED ALWAYS AS (
        to_tsvector('portuguese',
            coalesce(title, '') || ' ' || coalesce(summary, '') || ' ' || coalesce(body, ''))
    ) STORED,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_materias_search   ON materias USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_materias_tags      ON materias USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_materias_category  ON materias(category);
CREATE INDEX IF NOT EXISTS idx_materias_document  ON materias(document_id);
