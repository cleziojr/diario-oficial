package api

// Integration points (fora do escopo deste CRUD):
//   pdf-extraction → chama PATCH /api/v1/analyses/{id} para preencher extracted_text
//   ai-service     → chama POST /api/v1/documents/{documentId}/analyses com summary_text e insights

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type analysisStore interface {
	InsertAnalysis(ctx context.Context, arg sqlc.InsertAnalysisParams) (sqlc.DocumentAnalysis, error)
	GetAnalysisByID(ctx context.Context, id pgtype.UUID) (sqlc.DocumentAnalysis, error)
	ListAnalysesByDocumentID(ctx context.Context, documentID pgtype.UUID) ([]sqlc.DocumentAnalysis, error)
	UpdateAnalysis(ctx context.Context, arg sqlc.UpdateAnalysisParams) (sqlc.DocumentAnalysis, error)
	DeleteAnalysisByID(ctx context.Context, id pgtype.UUID) (int64, error)
}

type analysisJSON struct {
	ID            string          `json:"id"`
	DocumentID    string          `json:"document_id"`
	ExtractedText *string         `json:"extracted_text,omitempty"`
	SummaryText   string          `json:"summary_text"`
	Insights      json.RawMessage `json:"insights"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at,omitempty"`
}

type createAnalysisRequest struct {
	ExtractedText *string         `json:"extracted_text"`
	SummaryText   string          `json:"summary_text"`
	Insights      json.RawMessage `json:"insights"`
}

type updateAnalysisRequest struct {
	ExtractedText *string         `json:"extracted_text"`
	SummaryText   string          `json:"summary_text"`
	Insights      json.RawMessage `json:"insights"`
}

func analysisToJSON(a sqlc.DocumentAnalysis) (analysisJSON, error) {
	if !a.ID.Valid {
		return analysisJSON{}, errors.New("analysis id inválido")
	}
	idStr, err := formatPgUUID(a.ID)
	if err != nil {
		return analysisJSON{}, err
	}
	docIDStr, err := formatPgUUID(a.DocumentID)
	if err != nil {
		return analysisJSON{}, err
	}
	created := ""
	if a.CreatedAt.Valid {
		created = a.CreatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	updated := ""
	if a.UpdatedAt.Valid {
		updated = a.UpdatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	var extractedText *string
	if a.ExtractedText.Valid {
		extractedText = &a.ExtractedText.String
	}
	return analysisJSON{
		ID:            idStr,
		DocumentID:    docIDStr,
		ExtractedText: extractedText,
		SummaryText:   a.SummaryText,
		Insights:      a.Insights,
		CreatedAt:     created,
		UpdatedAt:     updated,
	}, nil
}

type analysisHandlers struct {
	q analysisStore
}

func (h *analysisHandlers) create(w http.ResponseWriter, r *http.Request) {
	docID, err := parseUUIDParam(chi.URLParam(r, "documentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documentId inválido")
		return
	}
	var body createAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	body.SummaryText = strings.TrimSpace(body.SummaryText)
	if body.SummaryText == "" {
		writeError(w, http.StatusBadRequest, "summary_text é obrigatório")
		return
	}
	if len(body.Insights) == 0 {
		writeError(w, http.StatusBadRequest, "insights é obrigatório")
		return
	}
	var extractedText pgtype.Text
	if body.ExtractedText != nil {
		extractedText = pgtype.Text{String: *body.ExtractedText, Valid: true}
	}
	analysis, err := h.q.InsertAnalysis(r.Context(), sqlc.InsertAnalysisParams{
		DocumentID:    docID,
		ExtractedText: extractedText,
		SummaryText:   body.SummaryText,
		Insights:      body.Insights,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusNotFound, "documento não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao salvar análise")
		return
	}
	out, err := analysisToJSON(analysis)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func (h *analysisHandlers) list(w http.ResponseWriter, r *http.Request) {
	docID, err := parseUUIDParam(chi.URLParam(r, "documentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documentId inválido")
		return
	}
	analyses, err := h.q.ListAnalysesByDocumentID(r.Context(), docID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar análises")
		return
	}
	items := make([]analysisJSON, 0, len(analyses))
	for _, a := range analyses {
		j, err := analysisToJSON(a)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		items = append(items, j)
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *analysisHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	analysis, err := h.q.GetAnalysisByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "análise não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar análise")
		return
	}
	out, err := analysisToJSON(analysis)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analysisHandlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	var body updateAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	body.SummaryText = strings.TrimSpace(body.SummaryText)
	if body.SummaryText == "" {
		writeError(w, http.StatusBadRequest, "summary_text é obrigatório")
		return
	}
	if len(body.Insights) == 0 {
		writeError(w, http.StatusBadRequest, "insights é obrigatório")
		return
	}
	var extractedText pgtype.Text
	if body.ExtractedText != nil {
		extractedText = pgtype.Text{String: *body.ExtractedText, Valid: true}
	}
	analysis, err := h.q.UpdateAnalysis(r.Context(), sqlc.UpdateAnalysisParams{
		ID:            id,
		ExtractedText: extractedText,
		SummaryText:   body.SummaryText,
		Insights:      body.Insights,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "análise não encontrada")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao atualizar análise")
		return
	}
	out, err := analysisToJSON(analysis)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *analysisHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	n, err := h.q.DeleteAnalysisByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover análise")
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "análise não encontrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mountDocumentAnalyses(r chi.Router, q analysisStore) {
	h := &analysisHandlers{q: q}
	r.Post("/", h.create)
	r.Get("/", h.list)
}

func mountAnalyses(r chi.Router, q analysisStore) {
	h := &analysisHandlers{q: q}
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.update)
	r.Delete("/{id}", h.delete)
}
