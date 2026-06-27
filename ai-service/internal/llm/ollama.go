package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"ai-service/internal/prompt"
)

var ollamaClient = &http.Client{Timeout: 5 * time.Minute}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func SummarizeOllama(ctx context.Context, text string) (string, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}

	reqBody := ollamaRequest{
		Model:  model,
		Prompt: prompt.Build(text),
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: serializar request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, host+"/api/generate", bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("ollama: criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ollamaClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: executar request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return "", fmt.Errorf("ollama: ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: retornou status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result ollamaResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("ollama: deserializar resposta: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("ollama: %s", result.Error)
	}
	if result.Response == "" {
		return "", fmt.Errorf("ollama: resposta vazia")
	}

	return result.Response, nil
}