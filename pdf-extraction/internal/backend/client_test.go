package backend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pdf-extraction/internal/backend"
)

func TestPostAnalysis_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("metodo inesperado: %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"summary_text": "Resumo de teste."}`))
	}))
	defer srv.Close()

	c := backend.New(srv.URL)
	summary, err := c.PostAnalysis(context.Background(), "doc-id-123", "texto extraido")
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if summary == "" {
		t.Fatal("resumo vazio")
	}
}

func TestPostAnalysis_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := backend.New(srv.URL)
	_, err := c.PostAnalysis(context.Background(), "doc-id-123", "texto")
	if err == nil {
		t.Fatal("esperava erro para status 500")
	}
}

func TestPostAnalysis_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := backend.New(srv.URL)
	_, err := c.PostAnalysis(ctx, "doc-id-123", "texto")
	if err == nil {
		t.Fatal("esperava erro de contexto cancelado")
	}
}
