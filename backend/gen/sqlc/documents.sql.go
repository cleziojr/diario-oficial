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

const listDocuments = `-- name: ListDocuments :many
SELECT id, filename, created_at
FROM documents
ORDER BY created_at DESC, id
LIMIT $1 OFFSET $2
`

type ListDocumentsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListDocuments(ctx context.Context, arg ListDocumentsParams) ([]Document, error) {
	rows, err := q.db.Query(ctx, listDocuments, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Document{}
	for rows.Next() {
		var i Document
		if err := rows.Scan(&i.ID, &i.Filename, &i.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const deleteDocumentByID = `-- name: DeleteDocumentByID :execrows
DELETE FROM documents
WHERE id = $1
`

func (q *Queries) DeleteDocumentByID(ctx context.Context, id pgtype.UUID) (int64, error) {
	result, err := q.db.Exec(ctx, deleteDocumentByID, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
