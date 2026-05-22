package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type documentInsightsStore interface {
	GetDocumentByID(ctx context.Context, id pgtype.UUID) (sqlc.Document, error)
	ListAnalysesByDocumentID(ctx context.Context, documentID pgtype.UUID) ([]sqlc.DocumentAnalysis, error)
}

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

func mountDocumentInsights(r chi.Router, q documentInsightsStore) {
	h := &documentInsightsHandlers{q: q}
	r.Get("/{id}/insights", h.get)
}
