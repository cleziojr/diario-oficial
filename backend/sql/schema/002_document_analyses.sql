CREATE TABLE document_analyses (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id    UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    extracted_text TEXT,
    summary_text   TEXT        NOT NULL,
    insights       JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ
);
