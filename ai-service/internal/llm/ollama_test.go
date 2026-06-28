package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ai-service/internal/llm"
)

func TestSummarizeOllama_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": "Resumo gerado com sucesso.",
		})
	}))
	defer srv.Close()

	os.Setenv("OLLAMA_HOST", srv.URL)
	os.Setenv("OLLAMA_MODEL", "llama3.2")

	result, err := llm.SummarizeOllama(context.Background(), "Texto de teste.")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result == "" {
		t.Fatal("resposta vazia")
	}
}

func TestSummarizeOllama_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": "",
		})
	}))
	defer srv.Close()

	os.Setenv("OLLAMA_HOST", srv.URL)

	_, err := llm.SummarizeOllama(context.Background(), "Texto de teste.")
	if err == nil {
		t.Fatal("esperava erro para resposta vazia")
	}
}

func TestSummarizeOllama_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	os.Setenv("OLLAMA_HOST", srv.URL)

	_, err := llm.SummarizeOllama(context.Background(), "Texto de teste.")
	if err == nil {
		t.Fatal("esperava erro para status 500")
	}
}

func TestSummarizeOllama_OllamaError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"response": "",
			"error":    "model not found",
		})
	}))
	defer srv.Close()

	os.Setenv("OLLAMA_HOST", srv.URL)

	_, err := llm.SummarizeOllama(context.Background(), "Texto de teste.")
	if err == nil {
		t.Fatal("esperava erro quando ollama retorna campo error")
	}
}