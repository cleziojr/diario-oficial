package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cleziojr/diario-oficial/backend/gen/sqlc"
	"github.com/cleziojr/diario-oficial/backend/pkg/aiclient"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubAnalysisStore struct {
	insertOut sqlc.DocumentAnalysis
	insertErr error

	getOut sqlc.DocumentAnalysis
	getErr error

	listOut []sqlc.DocumentAnalysis
	listErr error

	updateOut sqlc.DocumentAnalysis
	updateErr error

	delN   int64
	delErr error
}

func (s *stubAnalysisStore) InsertAnalysis(_ context.Context, _ sqlc.InsertAnalysisParams) (sqlc.DocumentAnalysis, error) {
	if s.insertErr != nil {
		return sqlc.DocumentAnalysis{}, s.insertErr
	}
	return s.insertOut, nil
}

func (s *stubAnalysisStore) GetAnalysisByID(_ context.Context, _ pgtype.UUID) (sqlc.DocumentAnalysis, error) {
	if s.getErr != nil {
		return sqlc.DocumentAnalysis{}, s.getErr
	}
	return s.getOut, nil
}

func (s *stubAnalysisStore) ListAnalysesByDocumentID(_ context.Context, _ pgtype.UUID) ([]sqlc.DocumentAnalysis, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listOut, nil
}

func (s *stubAnalysisStore) UpdateAnalysis(_ context.Context, _ sqlc.UpdateAnalysisParams) (sqlc.DocumentAnalysis, error) {
	if s.updateErr != nil {
		return sqlc.DocumentAnalysis{}, s.updateErr
	}
	return s.updateOut, nil
}

func (s *stubAnalysisStore) DeleteAnalysisByID(_ context.Context, _ pgtype.UUID) (int64, error) {
	if s.delErr != nil {
		return 0, s.delErr
	}
	return s.delN, nil
}

// mockAIServer sobe um httptest.Server que simula o ai-service.
func mockAIServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"summary": "resumo gerado",
			"model":   "mock",
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func testAnalysisRouter(t *testing.T, store analysisStore) http.Handler {
	t.Helper()
	ai := aiclient.New(mockAIServer(t).URL)
	r := chi.NewRouter()
	r.Route("/api/v1/documents/{documentId}/analyses", func(r chi.Router) {
		mountDocumentAnalyses(r, store, ai)
	})
	r.Route("/api/v1/analyses", func(r chi.Router) {
		mountAnalyses(r, store)
	})
	return r
}

func sampleAnalysis(t *testing.T) sqlc.DocumentAnalysis {
	t.Helper()
	ts := time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC)
	return sqlc.DocumentAnalysis{
		ID:          mustUUID(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"),
		DocumentID:  mustUUID(t, "550e8400-e29b-41d4-a716-446655440000"),
		SummaryText: "resumo do documento",
		Insights:    json.RawMessage(`[{"type":"keyword","value":"licitacao"}]`),
		CreatedAt:   pgtype.Timestamptz{Time: ts, Valid: true},
	}
}

const validDocID = "550e8400-e29b-41d4-a716-446655440000"
const validAnalysisID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"

func TestCreateAnalysis(t *testing.T) {
	store := &stubAnalysisStore{insertOut: sampleAnalysis(t)}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"summary_text":"resumo","insights":[{"type":"keyword"}]}`
	res, err := http.Post(srv.URL+"/api/v1/documents/"+validDocID+"/analyses", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got analysisJSON
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.SummaryText != "resumo do documento" || got.ID != validAnalysisID {
		t.Fatalf("resposta %+v", got)
	}
}

func TestCreateAnalysisWithExtractedText(t *testing.T) {
	store := &stubAnalysisStore{insertOut: sampleAnalysis(t)}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"extracted_text":"texto extraido do pdf","insights":[{"type":"keyword"}]}`
	res, err := http.Post(srv.URL+"/api/v1/documents/"+validDocID+"/analyses", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d — ai-service nao foi chamado corretamente", res.StatusCode)
	}
}

func TestCreateAnalysisDocumentNotFound(t *testing.T) {
	store := &stubAnalysisStore{insertErr: &pgconn.PgError{Code: "23503"}}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"summary_text":"resumo","insights":[{}]}`
	res, err := http.Post(srv.URL+"/api/v1/documents/"+validDocID+"/analyses", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestCreateAnalysisInvalidDocumentID(t *testing.T) {
	store := &stubAnalysisStore{}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"summary_text":"resumo","insights":[{}]}`
	res, err := http.Post(srv.URL+"/api/v1/documents/not-a-uuid/analyses", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestCreateAnalysisMissingSummary(t *testing.T) {
	store := &stubAnalysisStore{}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	// nem summary_text nem extracted_text preenchidos
	body := `{"summary_text":"  ","insights":[{}]}`
	res, err := http.Post(srv.URL+"/api/v1/documents/"+validDocID+"/analyses", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestGetAnalysisByID(t *testing.T) {
	store := &stubAnalysisStore{getOut: sampleAnalysis(t)}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/analyses/" + validAnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got analysisJSON
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.ID != validAnalysisID || got.DocumentID != validDocID {
		t.Fatalf("resposta %+v", got)
	}
}

func TestGetAnalysisNotFound(t *testing.T) {
	store := &stubAnalysisStore{getErr: pgx.ErrNoRows}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/analyses/" + validAnalysisID)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestListAnalysesByDocumentID(t *testing.T) {
	store := &stubAnalysisStore{listOut: []sqlc.DocumentAnalysis{sampleAnalysis(t)}}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/" + validDocID + "/analyses")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got []analysisJSON
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != validAnalysisID {
		t.Fatalf("resposta %+v", got)
	}
}

func TestUpdateAnalysis(t *testing.T) {
	tsUpd := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	out := sampleAnalysis(t)
	out.SummaryText = "novo resumo"
	out.UpdatedAt = pgtype.Timestamptz{Time: tsUpd, Valid: true}
	store := &stubAnalysisStore{updateOut: out}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"summary_text":"novo resumo","insights":[{"type":"updated"}]}`
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/analyses/"+validAnalysisID, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got analysisJSON
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.SummaryText != "novo resumo" || got.UpdatedAt == "" {
		t.Fatalf("resposta %+v", got)
	}
}

func TestUpdateAnalysisNotFound(t *testing.T) {
	store := &stubAnalysisStore{updateErr: pgx.ErrNoRows}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	body := `{"summary_text":"x","insights":[{}]}`
	req, err := http.NewRequest(http.MethodPatch, srv.URL+"/api/v1/analyses/"+validAnalysisID, bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestDeleteAnalysis(t *testing.T) {
	store := &stubAnalysisStore{delN: 1}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/analyses/"+validAnalysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestDeleteAnalysisNotFound(t *testing.T) {
	store := &stubAnalysisStore{delN: 0}
	srv := httptest.NewServer(testAnalysisRouter(t, store))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/analyses/"+validAnalysisID, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}