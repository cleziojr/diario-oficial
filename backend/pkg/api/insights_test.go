package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubDocumentInsightsStore struct {
	docOut sqlc.Document
	docErr error

	analysesOut []sqlc.DocumentAnalysis
	analysesErr error
}

func (s *stubDocumentInsightsStore) GetDocumentByID(_ context.Context, _ pgtype.UUID) (sqlc.Document, error) {
	if s.docErr != nil {
		return sqlc.Document{}, s.docErr
	}
	return s.docOut, nil
}

func (s *stubDocumentInsightsStore) ListAnalysesByDocumentID(_ context.Context, _ pgtype.UUID) ([]sqlc.DocumentAnalysis, error) {
	if s.analysesErr != nil {
		return nil, s.analysesErr
	}
	return s.analysesOut, nil
}

func testDocumentInsightsRouter(store documentInsightsStore) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/documents", func(r chi.Router) {
		mountDocumentInsights(r, store)
	})
	return r
}

func TestGetDocumentInsightsNestedJSON(t *testing.T) {
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	store := &stubDocumentInsightsStore{
		docOut: sqlc.Document{
			ID:        mustUUID(t, validDocID),
			Filename:  "diario.pdf",
			CreatedAt: pgtype.Timestamptz{Time: ts, Valid: true},
		},
		analysesOut: []sqlc.DocumentAnalysis{
			{
				ID:          mustUUID(t, validAnalysisID),
				DocumentID:  mustUUID(t, validDocID),
				SummaryText: "resumo 1",
				Insights:    json.RawMessage(`[{"type":"keyword","value":"licitação"},{"type":"deadline","value":"2026-06-01"}]`),
				CreatedAt:   pgtype.Timestamptz{Time: ts, Valid: true},
			},
			{
				ID:          mustUUID(t, "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"),
				DocumentID:  mustUUID(t, validDocID),
				SummaryText: "resumo 2",
				Insights:    json.RawMessage(`[{"type":"agency","value":"Prefeitura"}]`),
				CreatedAt:   pgtype.Timestamptz{Time: ts.Add(time.Minute), Valid: true},
			},
		},
	}
	srv := httptest.NewServer(testDocumentInsightsRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/" + validDocID + "/insights")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}

	var got documentInsightsResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Document.ID != validDocID || got.Document.Filename != "diario.pdf" {
		t.Fatalf("documento aninhado %+v", got.Document)
	}
	if len(got.Insights) != 3 {
		t.Fatalf("insights %+v", got.Insights)
	}
}

func TestGetDocumentInsightsNotFound(t *testing.T) {
	store := &stubDocumentInsightsStore{docErr: pgx.ErrNoRows}
	srv := httptest.NewServer(testDocumentInsightsRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/" + validDocID + "/insights")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestGetDocumentInsightsInvalidID(t *testing.T) {
	store := &stubDocumentInsightsStore{}
	srv := httptest.NewServer(testDocumentInsightsRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/not-a-uuid/insights")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}
