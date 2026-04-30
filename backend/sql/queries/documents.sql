-- name: InsertDocument :one
INSERT INTO documents (filename)
VALUES ($1)
RETURNING id, filename, created_at;

-- name: GetDocumentByID :one
SELECT id, filename, created_at
FROM documents
WHERE id = $1;

-- name: ListDocuments :many
SELECT id, filename, created_at
FROM documents
ORDER BY created_at DESC, id
LIMIT $1 OFFSET $2;

-- name: DeleteDocumentByID :execrows
DELETE FROM documents
WHERE id = $1;
