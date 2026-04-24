-- name: InsertDocument :one
INSERT INTO documents (filename)
VALUES ($1)
RETURNING id, filename, created_at;

-- name: GetDocumentByID :one
SELECT id, filename, created_at
FROM documents
WHERE id = $1;
