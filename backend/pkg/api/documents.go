package api

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

type documentStore interface {
	InsertDocument(ctx context.Context, filename string) (sqlc.Document, error)
	GetDocumentByID(ctx context.Context, id pgtype.UUID) (sqlc.Document, error)
	ListDocuments(ctx context.Context, arg sqlc.ListDocumentsParams) ([]sqlc.Document, error)
	DeleteDocumentByID(ctx context.Context, id pgtype.UUID) (int64, error)
}

type documentJSON struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	CreatedAt string `json:"created_at"`
}

type listDocumentsResponse struct {
	Items []documentJSON `json:"items"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type createDocumentRequest struct {
	Filename string `json:"filename"`
}

func documentToJSON(d sqlc.Document) (documentJSON, error) {
	if !d.ID.Valid {
		return documentJSON{}, errors.New("document id inválido")
	}
	idStr, err := formatPgUUID(d.ID)
	if err != nil {
		return documentJSON{}, err
	}
	created := ""
	if d.CreatedAt.Valid {
		created = d.CreatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	return documentJSON{
		ID:        idStr,
		Filename:  d.Filename,
		CreatedAt: created,
	}, nil
}

func formatPgUUID(u pgtype.UUID) (string, error) {
	if !u.Valid {
		return "", errors.New("uuid inválido")
	}
	b := u.Bytes[:]
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(b[0:4]),
		binary.BigEndian.Uint16(b[4:6]),
		binary.BigEndian.Uint16(b[6:8]),
		binary.BigEndian.Uint16(b[8:10]),
		uint64(b[10])<<40|uint64(b[11])<<32|uint64(b[12])<<24|uint64(b[13])<<16|uint64(b[14])<<8|uint64(b[15]),
	), nil
}

func parseUUIDParam(s string) (pgtype.UUID, error) {
	s = strings.ReplaceAll(strings.TrimSpace(s), "-", "")
	if len(s) != 32 {
		return pgtype.UUID{}, fmt.Errorf("uuid inválido")
	}
	var b [16]byte
	if _, err := hex.Decode(b[:], []byte(s)); err != nil {
		return pgtype.UUID{}, fmt.Errorf("uuid inválido")
	}
	return pgtype.UUID{Bytes: b, Valid: true}, nil
}

func parsePagination(r *http.Request) (page, limit int, offset int32, err error) {
	page = defaultPage
	limit = defaultLimit
	if v := r.URL.Query().Get("page"); v != "" {
		p, e := strconv.Atoi(v)
		if e != nil || p < 1 {
			return 0, 0, 0, fmt.Errorf("page inválido")
		}
		page = p
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		l, e := strconv.Atoi(v)
		if e != nil || l < 1 {
			return 0, 0, 0, fmt.Errorf("limit inválido")
		}
		if l > maxLimit {
			l = maxLimit
		}
		limit = l
	}
	off := (page - 1) * limit
	if off > int(^uint32(0)>>1) {
		return 0, 0, 0, fmt.Errorf("paginação fora do intervalo")
	}
	return page, limit, int32(off), nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

type documentHandlers struct {
	q documentStore
}

func (h *documentHandlers) create(w http.ResponseWriter, r *http.Request) {
	var body createDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	body.Filename = strings.TrimSpace(body.Filename)
	if body.Filename == "" {
		writeError(w, http.StatusBadRequest, "filename é obrigatório")
		return
	}
	doc, err := h.q.InsertDocument(r.Context(), body.Filename)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao salvar documento")
		return
	}
	out, err := documentToJSON(doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *documentHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	doc, err := h.q.GetDocumentByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "documento não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar documento")
		return
	}
	out, err := documentToJSON(doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *documentHandlers) list(w http.ResponseWriter, r *http.Request) {
	page, limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	docs, err := h.q.ListDocuments(r.Context(), sqlc.ListDocumentsParams{
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar documentos")
		return
	}
	items := make([]documentJSON, 0, len(docs))
	for _, d := range docs {
		j, err := documentToJSON(d)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		items = append(items, j)
	}
	writeJSON(w, http.StatusOK, listDocumentsResponse{
		Items: items,
		Page:  page,
		Limit: limit,
	})
}

func (h *documentHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	n, err := h.q.DeleteDocumentByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover documento")
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "documento não encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mountDocuments(r chi.Router, q documentStore) {
	h := &documentHandlers{q: q}
	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{id}", h.get)
	r.Delete("/{id}", h.delete)
}
