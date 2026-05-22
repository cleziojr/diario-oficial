package api

// Insights são gerados por modelos LLM a partir de uma análise de documento do Diário Oficial.
// Relação: documents 1→N document_analyses 1→N insights
// O campo `model` permite comparar respostas de LLMs diferentes sobre o mesmo documento.

import (
	"bytes"
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

// ---------------------------------------------------------------------------
// Store interfaces
// ---------------------------------------------------------------------------

type insightStore interface {
	InsertInsight(ctx context.Context, arg sqlc.InsertInsightParams) (sqlc.Insight, error)
	GetInsightByID(ctx context.Context, id pgtype.UUID) (sqlc.Insight, error)
	ListInsightsByDocumentID(ctx context.Context, documentID pgtype.UUID) ([]sqlc.Insight, error)
	ListInsightsByAnalysisID(ctx context.Context, analysisID pgtype.UUID) ([]sqlc.Insight, error)
	ListInsightsByDocumentIDAndModel(ctx context.Context, arg sqlc.ListInsightsByDocumentIDAndModelParams) ([]sqlc.Insight, error)
	UpdateInsight(ctx context.Context, arg sqlc.UpdateInsightParams) (sqlc.Insight, error)
	DeleteInsightByID(ctx context.Context, id pgtype.UUID) (int64, error)
	DeleteInsightsByDocumentID(ctx context.Context, documentID pgtype.UUID) (int64, error)
	DeleteInsightsByAnalysisID(ctx context.Context, analysisID pgtype.UUID) (int64, error)
}

// documentInsightsStore é usado apenas pelo agregador legacy.
type documentInsightsStore interface {
	GetDocumentByID(ctx context.Context, id pgtype.UUID) (sqlc.Document, error)
	ListAnalysesByDocumentID(ctx context.Context, documentID pgtype.UUID) ([]sqlc.DocumentAnalysis, error)
}

// ---------------------------------------------------------------------------
// JSON types
// ---------------------------------------------------------------------------

type insightJSON struct {
	ID         string          `json:"id"`
	DocumentID string          `json:"document_id"`
	AnalysisID string          `json:"analysis_id"`
	Model      string          `json:"model"`
	Content    string          `json:"content"`
	Metadata   json.RawMessage `json:"metadata"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
}

type createInsightRequest struct {
	AnalysisID string          `json:"analysis_id"`
	Model      string          `json:"model"`
	Content    string          `json:"content"`
	Metadata   json.RawMessage `json:"metadata"`
}

type updateInsightRequest struct {
	Content  string          `json:"content"`
	Metadata json.RawMessage `json:"metadata"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func insightToJSON(i sqlc.Insight) (insightJSON, error) {
	if !i.ID.Valid {
		return insightJSON{}, errors.New("insight id inválido")
	}
	idStr, err := formatPgUUID(i.ID)
	if err != nil {
		return insightJSON{}, err
	}
	docIDStr, err := formatPgUUID(i.DocumentID)
	if err != nil {
		return insightJSON{}, err
	}
	analysisIDStr, err := formatPgUUID(i.AnalysisID)
	if err != nil {
		return insightJSON{}, err
	}
	created := ""
	if i.CreatedAt.Valid {
		created = i.CreatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	updated := ""
	if i.UpdatedAt.Valid {
		updated = i.UpdatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	metadata := i.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}
	return insightJSON{
		ID:         idStr,
		DocumentID: docIDStr,
		AnalysisID: analysisIDStr,
		Model:      i.Model,
		Content:    i.Content,
		Metadata:   metadata,
		CreatedAt:  created,
		UpdatedAt:  updated,
	}, nil
}

// appendAnalysisInsights achata o campo JSONB legacy de document_analyses.
func appendAnalysisInsights(items []json.RawMessage, raw json.RawMessage) ([]json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return items, nil
	}
	if trimmed[0] != '[' {
		if !json.Valid(trimmed) {
			return nil, errors.New("insights inválido")
		}
		return append(items, json.RawMessage(append([]byte(nil), trimmed...))), nil
	}
	var current []json.RawMessage
	if err := json.Unmarshal(trimmed, &current); err != nil {
		return nil, err
	}
	return append(items, current...), nil
}

// ---------------------------------------------------------------------------
// CRUD handlers — tabela insights normalizada
// ---------------------------------------------------------------------------

type insightHandlers struct {
	q insightStore
}

// POST /api/v1/documents/{documentId}/insights
func (h *insightHandlers) create(w http.ResponseWriter, r *http.Request) {
	docID, err := parseUUIDParam(chi.URLParam(r, "documentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documentId inválido")
		return
	}
	var body createInsightRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	body.Model = strings.TrimSpace(body.Model)
	if body.Model == "" {
		writeError(w, http.StatusBadRequest, "model é obrigatório")
		return
	}
	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		writeError(w, http.StatusBadRequest, "content é obrigatório")
		return
	}
	analysisID, err := parseUUIDParam(body.AnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "analysis_id inválido")
		return
	}
	metadata := body.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	insight, err := h.q.InsertInsight(r.Context(), sqlc.InsertInsightParams{
		DocumentID: docID,
		AnalysisID: analysisID,
		Model:      body.Model,
		Content:    body.Content,
		Metadata:   metadata,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			writeError(w, http.StatusNotFound, "documento ou análise não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao salvar insight")
		return
	}
	out, err := insightToJSON(insight)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// GET /api/v1/documents/{documentId}/insights?model=gpt-4o
func (h *insightHandlers) listByDocument(w http.ResponseWriter, r *http.Request) {
	docID, err := parseUUIDParam(chi.URLParam(r, "documentId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documentId inválido")
		return
	}
	if model := r.URL.Query().Get("model"); model != "" {
		insights, err := h.q.ListInsightsByDocumentIDAndModel(r.Context(), sqlc.ListInsightsByDocumentIDAndModelParams{
			DocumentID: docID,
			Model:      model,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao listar insights")
			return
		}
		h.writeList(w, insights)
		return
	}
	insights, err := h.q.ListInsightsByDocumentID(r.Context(), docID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar insights")
		return
	}
	h.writeList(w, insights)
}

// GET /api/v1/analyses/{analysisId}/insights
func (h *insightHandlers) listByAnalysis(w http.ResponseWriter, r *http.Request) {
	analysisID, err := parseUUIDParam(chi.URLParam(r, "analysisId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "analysisId inválido")
		return
	}
	insights, err := h.q.ListInsightsByAnalysisID(r.Context(), analysisID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar insights")
		return
	}
	h.writeList(w, insights)
}

// GET /api/v1/insights/{id}
func (h *insightHandlers) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	insight, err := h.q.GetInsightByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "insight não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao buscar insight")
		return
	}
	out, err := insightToJSON(insight)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// PATCH /api/v1/insights/{id}
func (h *insightHandlers) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	var body updateInsightRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON inválido")
		return
	}
	body.Content = strings.TrimSpace(body.Content)
	if body.Content == "" {
		writeError(w, http.StatusBadRequest, "content é obrigatório")
		return
	}
	metadata := body.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("{}")
	}
	insight, err := h.q.UpdateInsight(r.Context(), sqlc.UpdateInsightParams{
		ID:       id,
		Content:  body.Content,
		Metadata: metadata,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusNotFound, "insight não encontrado")
			return
		}
		writeError(w, http.StatusInternalServerError, "erro ao atualizar insight")
		return
	}
	out, err := insightToJSON(insight)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// DELETE /api/v1/insights/{id}
func (h *insightHandlers) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	n, err := h.q.DeleteInsightByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao remover insight")
		return
	}
	if n == 0 {
		writeError(w, http.StatusNotFound, "insight não encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *insightHandlers) writeList(w http.ResponseWriter, insights []sqlc.Insight) {
	items := make([]insightJSON, 0, len(insights))
	for _, i := range insights {
		j, err := insightToJSON(i)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
			return
		}
		items = append(items, j)
	}
	writeJSON(w, http.StatusOK, items)
}

// ---------------------------------------------------------------------------
// Agregador legacy — GET /documents/{id}/insights via JSONB de document_analyses
// ---------------------------------------------------------------------------

type documentInsightsResponse struct {
	Document documentJSON      `json:"document"`
	Insights []json.RawMessage `json:"insights"`
}

type documentInsightsHandlers struct {
	q documentInsightsStore
}

func (h *documentInsightsHandlers) get(w http.ResponseWriter, r *http.Request) {
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
	analyses, err := h.q.ListAnalysesByDocumentID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao listar análises")
		return
	}
	document, err := documentToJSON(doc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro ao montar resposta")
		return
	}
	insights := make([]json.RawMessage, 0)
	for _, analysis := range analyses {
		insights, err = appendAnalysisInsights(insights, analysis.Insights)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro ao montar insights")
			return
		}
	}
	writeJSON(w, http.StatusOK, documentInsightsResponse{
		Document: document,
		Insights: insights,
	})
}

// ---------------------------------------------------------------------------
// Mount functions
// ---------------------------------------------------------------------------

// mountDocumentInsights registra o agregador legacy GET /documents/{id}/insights.
// Recebe documentInsightsStore — compatível com o teste existente.
func mountDocumentInsights(r chi.Router, q documentInsightsStore) {
	h := &documentInsightsHandlers{q: q}
	r.Get("/{id}/insights", h.get)
}

// mountDocumentInsightsCRUD registra POST/GET /documents/{documentId}/insights.
// Chamado separadamente no router com sqlc.Queries que implementa insightStore.
func mountDocumentInsightsCRUD(r chi.Router, q insightStore) {
	h := &insightHandlers{q: q}
	r.Route("/{documentId}/insights", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.listByDocument)
	})
}

// mountAnalysisInsights registra GET /analyses/{analysisId}/insights
func mountAnalysisInsights(r chi.Router, q insightStore) {
	h := &insightHandlers{q: q}
	r.Get("/{analysisId}/insights", h.listByAnalysis)
}

// mountInsights registra GET/PATCH/DELETE /insights/{id}
func mountInsights(r chi.Router, q insightStore) {
	h := &insightHandlers{q: q}
	r.Get("/{id}", h.get)
	r.Patch("/{id}", h.update)
	r.Delete("/{id}", h.delete)
}