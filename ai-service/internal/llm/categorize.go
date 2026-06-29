package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"ai-service/internal/prompt"
)

// CategorizeOpenRouter pede ao OpenRouter a categorização do trecho e devolve a
// resposta bruta (espera-se um JSON conforme prompt.BuildCategorize).
func CategorizeOpenRouter(ctx context.Context, text string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY nao definida")
	}
	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}

	reqBody := map[string]any{
		"model":           model,
		"stream":          false,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []HFMessage{
			{Role: "user", Content: prompt.BuildCategorize(text)},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("serializar request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, APIURL, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/cleziojr/diario-oficial")
	req.Header.Set("X-Title", "Insight Diario")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executar request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", fmt.Errorf("ler resposta: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var partial HFResponse
		if json.Unmarshal(bodyBytes, &partial) == nil && partial.Error != nil {
			return "", fmt.Errorf("API retornou %d: %s", resp.StatusCode, partial.Error.Message)
		}
		return "", fmt.Errorf("API retornou status %d", resp.StatusCode)
	}

	var result HFResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("deserializar resposta: %w", err)
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("resposta vazia da API")
	}
	return result.Choices[0].Message.Content, nil
}

// CategorizeOllama faz o mesmo via Ollama local, forçando saída em JSON.
func CategorizeOllama(ctx context.Context, text string) (string, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3.2"
	}

	reqBody := map[string]any{
		"model":  model,
		"prompt": prompt.BuildCategorize(text),
		"stream": false,
		"format": "json",
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

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
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
	return result.Response, nil
}
