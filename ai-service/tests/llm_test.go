package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ai-service/internal/llm"
)

func mockServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
}

func TestSummarize_Success(t *testing.T) {
	srv := mockServer(http.StatusOK, `{"choices":[{"message":{"role":"assistant","content":"1. Resumo\n2. Mock ok"}}]}`)
	defer srv.Close()

	llm.APIURL = srv.URL
	os.Setenv("OPENROUTER_API_KEY", "chave-fake")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	summary, err := llm.Summarize(context.Background(), "Texto de exemplo")
	if err != nil {
		t.Fatalf("nao esperava erro: %v", err)
	}
	if summary != "1. Resumo\n2. Mock ok" {
		t.Errorf("resumo inesperado: %q", summary)
	}
}

func TestSummarize_MissingAPIKey(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")
	_, err := llm.Summarize(context.Background(), "Texto")
	if err == nil {
		t.Error("esperava erro por falta de API key")
	}
}

func TestSummarize_APIError(t *testing.T) {
	srv := mockServer(http.StatusInternalServerError, `{"error":{"message":"rate limit"}}`)
	defer srv.Close()

	llm.APIURL = srv.URL
	os.Setenv("OPENROUTER_API_KEY", "chave-fake")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	_, err := llm.Summarize(context.Background(), "Texto")
	if err == nil {
		t.Error("esperava erro para status 500")
	}
}

func TestSummarize_EmptyChoices(t *testing.T) {
	srv := mockServer(http.StatusOK, `{"choices":[]}`)
	defer srv.Close()

	llm.APIURL = srv.URL
	os.Setenv("OPENROUTER_API_KEY", "chave-fake")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	_, err := llm.Summarize(context.Background(), "Texto")
	if err == nil {
		t.Error("esperava erro para choices vazio")
	}
}

func TestSummarize_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	llm.APIURL = srv.URL
	os.Setenv("OPENROUTER_API_KEY", "chave-fake")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := llm.Summarize(ctx, "Texto")
	if err == nil {
		t.Error("esperava erro de contexto cancelado")
	}
}