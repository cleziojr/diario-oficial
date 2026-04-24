package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"ai-service/internal/llm"
)

func TestSummarize_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [
				{"message": {"role": "assistant", "content": "1. Teste de resumo\n2. Mock funcionando"}}
			]
		}`))
	}))
	defer mockServer.Close()

	originalURL := llm.APIURL
	llm.APIURL = mockServer.URL
	defer func() { llm.APIURL = originalURL }()

	os.Setenv("OPENROUTER_API_KEY", "chave-fake")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	summary, err := llm.Summarize("Texto de exemplo")

	if err != nil {
		t.Fatalf("Não esperava erro, mas obteve: %v", err)
	}

	expected := "1. Teste de resumo\n2. Mock funcionando"
	if summary != expected {
		t.Errorf("Esperava o resumo %q, obteve %q", expected, summary)
	}
}

func TestSummarize_MissingAPIKey(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")

	_, err := llm.Summarize("Texto")

	if err == nil {
		t.Error("Esperava um erro por falta de API KEY, mas a função retornou sucesso")
	}
}
