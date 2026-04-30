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
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubDocumentStore struct {
	insertOut sqlc.Document
	insertErr error

	getOut sqlc.Document
	getErr error

	listOut []sqlc.Document
	listErr error

	delN   int64
	delErr error
}

func (s *stubDocumentStore) InsertDocument(ctx context.Context, filename string) (sqlc.Document, error) {
	if s.insertErr != nil {
		return sqlc.Document{}, s.insertErr
	}
	return s.insertOut, nil
}

func (s *stubDocumentStore) GetDocumentByID(ctx context.Context, id pgtype.UUID) (sqlc.Document, error) {
	if s.getErr != nil {
		return sqlc.Document{}, s.getErr
	}
	return s.getOut, nil
}

func (s *stubDocumentStore) ListDocuments(ctx context.Context, arg sqlc.ListDocumentsParams) ([]sqlc.Document, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	_ = arg
	return s.listOut, nil
}

func (s *stubDocumentStore) DeleteDocumentByID(ctx context.Context, id pgtype.UUID) (int64, error) {
	if s.delErr != nil {
		return 0, s.delErr
	}
	return s.delN, nil
}

func testDocRouter(store documentStore) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1/documents", func(r chi.Router) {
		mountDocuments(r, store)
	})
	return r
}

func mustUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	u, err := parseUUIDParam(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestCreateDocument(t *testing.T) {
	id := mustUUID(t, "550e8400-e29b-41d4-a716-446655440000")
	ts := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	store := &stubDocumentStore{
		insertOut: sqlc.Document{
			ID:        id,
			Filename:  "relatorio.pdf",
			CreatedAt: pgtype.Timestamptz{Time: ts, Valid: true},
		},
	}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	body := `{"filename":"relatorio.pdf"}`
	res, err := http.Post(srv.URL+"/api/v1/documents", "application/json", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got documentJSON
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Filename != "relatorio.pdf" || got.ID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("resposta %+v", got)
	}
}

func TestGetDocumentInvalidID(t *testing.T) {
	store := &stubDocumentStore{}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/not-a-uuid")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestGetDocumentNotFound(t *testing.T) {
	store := &stubDocumentStore{getErr: pgx.ErrNoRows}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents/550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestDeleteDocumentNotFound(t *testing.T) {
	store := &stubDocumentStore{delN: 0}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/documents/550e8400-e29b-41d4-a716-446655440000", nil)
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

func TestDeleteDocumentNoContent(t *testing.T) {
	store := &stubDocumentStore{delN: 1}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/documents/550e8400-e29b-41d4-a716-446655440000", nil)
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

func TestListDocuments(t *testing.T) {
	id := mustUUID(t, "550e8400-e29b-41d4-a716-446655440000")
	ts := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	store := &stubDocumentStore{
		listOut: []sqlc.Document{{
			ID:        id,
			Filename:  "a.pdf",
			CreatedAt: pgtype.Timestamptz{Time: ts, Valid: true},
		}},
	}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Get(srv.URL + "/api/v1/documents?page=2&limit=10")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	var got listDocumentsResponse
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Page != 2 || got.Limit != 10 || len(got.Items) != 1 {
		t.Fatalf("resposta %+v", got)
	}
}

func TestCreateDocumentEmptyFilename(t *testing.T) {
	store := &stubDocumentStore{}
	srv := httptest.NewServer(testDocRouter(store))
	t.Cleanup(srv.Close)

	res, err := http.Post(srv.URL+"/api/v1/documents", "application/json", bytes.NewReader([]byte(`{"filename":"  "}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status %d", res.StatusCode)
	}
}
