package tests

import (
	"testing"

	"ai-service/internal/llm"
)

func TestSummarize(t *testing.T) {
	text := "Teste simples de texto"

	_, err := llm.Summarize(text)

	if err != nil {
		t.Log("Erro esperado se API key não estiver definida:", err)
	}
}