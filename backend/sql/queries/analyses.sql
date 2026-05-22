-- name: InsertAnalysis :one
INSERT INTO document_analyses (document_id, extracted_text, summary_text, insights)
VALUES ($1, $2, $3, $4)
RETURNING id, document_id, extracted_text, summary_text, insights, created_at, updated_at;

-- name: GetAnalysisByID :one
SELECT id, document_id, extracted_text, summary_text, insights, created_at, updated_at
FROM document_analyses
WHERE id = $1;

-- name: ListAnalysesByDocumentID :many
SELECT id, document_id, extracted_text, summary_text, insights, created_at, updated_at
FROM document_analyses
WHERE document_id = $1
ORDER BY created_at DESC;

-- name: UpdateAnalysis :one
UPDATE document_analyses
SET extracted_text = $2, summary_text = $3, insights = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, document_id, extracted_text, summary_text, insights, created_at, updated_at;

-- name: DeleteAnalysisByID :execrows
DELETE FROM document_analyses
WHERE id = $1;
