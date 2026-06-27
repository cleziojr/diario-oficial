package aiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client chama o ai-service via HTTP.
type Client struct {
	baseURL string
	http    *http.Client
}

// New cria um Client apontando para baseURL (ex: "http://ai-service:9090").
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Minute},
	}
}

type summarizeRequest struct {
	Text string `json:"text"`
}

// SummarizeResponse e a resposta do ai-service.
type SummarizeResponse struct {
	Summary string `json:"summary"`
	Model   string `json:"model"`
}

// Summarize envia o texto ao ai-service e retorna o resumo gerado.
func (c *Client) Summarize(ctx context.Context, text string) (*SummarizeResponse, error) {
	body, err := json.Marshal(summarizeRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("aiclient: serializar request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/summarize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("aiclient: criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aiclient: executar request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("aiclient: ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(raw, &errBody) == nil && errBody.Error != "" {
			return nil, fmt.Errorf("aiclient: ai-service retornou %d: %s", resp.StatusCode, errBody.Error)
		}
		return nil, fmt.Errorf("aiclient: ai-service retornou status %d", resp.StatusCode)
	}

	var out SummarizeResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("aiclient: deserializar resposta: %w", err)
	}

	return &out, nil
}
