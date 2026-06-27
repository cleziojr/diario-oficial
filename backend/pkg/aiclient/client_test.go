package aiclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cleziojr/diario-oficial/backend/pkg/aiclient"
)

func mockAIServer(t *testing.T, status int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/summarize" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
}

func TestSummarize_Success(t *testing.T) {
	srv := mockAIServer(t, http.StatusOK, map[string]string{
		"summary": "resumo gerado",
		"model":   "ollama(llama3.2)",
	})
	defer srv.Close()

	c := aiclient.New(srv.URL)
	res, err := c.Summarize(context.Background(), "Texto de exemplo")
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if res.Summary != "resumo gerado" {
		t.Errorf("summary inesperado: %q", res.Summary)
	}
	if res.Model != "ollama(llama3.2)" {
		t.Errorf("model inesperado: %q", res.Model)
	}
}

func TestSummarize_ServiceError(t *testing.T) {
	srv := mockAIServer(t, http.StatusBadGateway, map[string]string{
		"error": "timeout ao chamar ollama",
	})
	defer srv.Close()

	c := aiclient.New(srv.URL)
	_, err := c.Summarize(context.Background(), "Texto")
	if err == nil {
		t.Fatal("esperava erro para status 502")
	}
}

func TestSummarize_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := aiclient.New(srv.URL)
	_, err := c.Summarize(ctx, "Texto")
	if err == nil {
		t.Fatal("esperava erro de contexto cancelado")
	}
}
