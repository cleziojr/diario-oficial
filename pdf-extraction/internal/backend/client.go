package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client chama o backend para persistir o texto extraído de um PDF.
type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type createAnalysisRequest struct {
	ExtractedText string          `json:"extracted_text"`
	SummaryText   string          `json:"summary_text"`
	Insights      json.RawMessage `json:"insights"`
}

// PostAnalysis envia o texto extraído ao backend para um documento existente.
// O backend se encarrega de chamar o ai-service e gerar o summary_text.
func (c *Client) PostAnalysis(ctx context.Context, documentID, extractedText string) (string, error) {
	body, err := json.Marshal(createAnalysisRequest{
		ExtractedText: extractedText,
		SummaryText:   "-",
		Insights:      json.RawMessage(`[]`),
	})
	if err != nil {
		return "", fmt.Errorf("backend client: serializar request: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/documents/%s/analyses", c.baseURL, documentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("backend client: criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("backend client: executar request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("backend client: ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("backend client: status inesperado %d", resp.StatusCode)
	}

	var out struct {
		SummaryText string `json:"summary_text"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", fmt.Errorf("backend client: deserializar resposta: %w", err)
	}

	return out.SummaryText, nil
}
