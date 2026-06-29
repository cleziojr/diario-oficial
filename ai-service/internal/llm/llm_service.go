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

var APIURL = "https://openrouter.ai/api/v1/chat/completions"

type HFRequest struct {
	Model    string      `json:"model"`
	Messages []HFMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

type HFMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type HFResponse struct {
	Choices []struct {
		Message HFMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

var client = &http.Client{Timeout: 60 * time.Second}

func Summarize(ctx context.Context, text string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENROUTER_API_KEY nao definida")
	}

	model := os.Getenv("OPENROUTER_MODEL")
	if model == "" {
		model = "meta-llama/llama-3.3-70b-instruct:free"
	}

	reqBody := HFRequest{
		Model:  model,
		Stream: false,
		Messages: []HFMessage{
			{Role: "user", Content: prompt.Build(text)},
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

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
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