-- Normaliza o campo insights JSONB de document_analyses
-- para uma relação 1:N, permitindo múltiplos modelos LLM por documento

CREATE TABLE insights (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- FK para o documento do Diário Oficial que originou o insight
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,

    -- FK para a análise que gerou este insight (texto extraído + summary)
    analysis_id UUID        NOT NULL REFERENCES document_analyses(id) ON DELETE CASCADE,

    -- Modelo LLM que gerou o insight (ex: "gpt-4o", "gemini-1.5-pro", "claude-3-5-sonnet")
    model       TEXT        NOT NULL,

    -- Resposta bruta do modelo
    content     TEXT        NOT NULL,

    -- Metadados de controle: tokens usados, latência, versão do prompt, etc.
    metadata    JSONB       NOT NULL DEFAULT '{}',

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_insights_document_id ON insights(document_id);
CREATE INDEX idx_insights_analysis_id ON insights(analysis_id);
CREATE INDEX idx_insights_model       ON insights(model);
CREATE INDEX idx_insights_metadata    ON insights USING GIN(metadata);