package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getDocumentByID = `-- name: GetDocumentByID :one
SELECT id, filename, created_at
FROM documents
WHERE id = $1
`

func (q *Queries) GetDocumentByID(ctx context.Context, id pgtype.UUID) (Document, error) {
	row := q.db.QueryRow(ctx, getDocumentByID, id)
	var i Document
	err := row.Scan(&i.ID, &i.Filename, &i.CreatedAt)
	return i, err
}

const insertDocument = `-- name: InsertDocument :one
INSERT INTO documents (filename)
VALUES ($1)
RETURNING id, filename, created_at
`

func (q *Queries) InsertDocument(ctx context.Context, filename string) (Document, error) {
	row := q.db.QueryRow(ctx, insertDocument, filename)
	var i Document
	err := row.Scan(&i.ID, &i.Filename, &i.CreatedAt)
	return i, err
}
