-- name: InsertInsight :one
INSERT INTO insights (document_id, analysis_id, model, content, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, document_id, analysis_id, model, content, metadata, created_at, updated_at;

-- name: GetInsightByID :one
SELECT id, document_id, analysis_id, model, content, metadata, created_at, updated_at
FROM insights
WHERE id = $1;

-- name: ListInsightsByDocumentID :many
-- Todos os insights de um documento, de todos os modelos
SELECT id, document_id, analysis_id, model, content, metadata, created_at, updated_at
FROM insights
WHERE document_id = $1
ORDER BY model, created_at DESC;

-- name: ListInsightsByAnalysisID :many
-- Todos os insights gerados a partir de uma análise específica
SELECT id, document_id, analysis_id, model, content, metadata, created_at, updated_at
FROM insights
WHERE analysis_id = $1
ORDER BY model, created_at DESC;

-- name: ListInsightsByModel :many
-- Todos os insights gerados por um modelo específico
SELECT id, document_id, analysis_id, model, content, metadata, created_at, updated_at
FROM insights
WHERE model = $1
ORDER BY created_at DESC;

-- name: ListInsightsByDocumentIDAndModel :many
-- Insights de um documento filtrados por modelo (comparar respostas de LLMs diferentes)
SELECT id, document_id, analysis_id, model, content, metadata, created_at, updated_at
FROM insights
WHERE document_id = $1
  AND model = $2
ORDER BY created_at DESC;

-- name: UpdateInsight :one
UPDATE insights
SET content    = $2,
    metadata   = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, document_id, analysis_id, model, content, metadata, created_at, updated_at;

-- name: DeleteInsightByID :execrows
DELETE FROM insights
WHERE id = $1;

-- name: DeleteInsightsByDocumentID :execrows
DELETE FROM insights
WHERE document_id = $1;

-- name: DeleteInsightsByAnalysisID :execrows
DELETE FROM insights
WHERE analysis_id = $1;