package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Document struct {
	ID        pgtype.UUID        `json:"id"`
	Filename  string             `json:"filename"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}
