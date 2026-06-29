package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-service/internal/api"
)

// --- mock do provider ---

type stubProvider struct {
	summary string
	err     error
}

func (s *stubProvider) Summarize(_ context.Context, _ string) (string, error) {
	return s.summary, s.err
}
func (s *stubProvider) Name() string { return "stub" }

// --- helpers ---

func newServer(t *testing.T, p *stubProvider) *httptest.Server {
	t.Helper()
	return httptest.NewServer(api.NewRouter(p))
}

func postSummarize(t *testing.T, srv *httptest.Server, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(srv.URL+"/summarize", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("request falhou: %v", err)
	}
	return resp
}

// --- testes ---

func TestSummarize_Success(t *testing.T) {
	srv := newServer(t, &stubProvider{summary: "resumo gerado"})
	defer srv.Close()

	resp := postSummarize(t, srv, `{"text":"Texto de exemplo"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", resp.StatusCode)
	}

	var out struct {
		Summary string `json:"summary"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode falhou: %v", err)
	}
	if out.Summary != "resumo gerado" {
		t.Errorf("summary inesperado: %q", out.Summary)
	}
	if out.Model != "stub" {
		t.Errorf("model inesperado: %q", out.Model)
	}
}

func TestSummarize_MissingText(t *testing.T) {
	srv := newServer(t, &stubProvider{})
	defer srv.Close()

	for _, body := range []string{`{}`, `{"text":""}`, `{"text":"   "}`} {
		resp := postSummarize(t, srv, body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("body %q: esperava 400, obteve %d", body, resp.StatusCode)
		}
	}
}

func TestSummarize_InvalidJSON(t *testing.T) {
	srv := newServer(t, &stubProvider{})
	defer srv.Close()

	resp := postSummarize(t, srv, `nao e json`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", resp.StatusCode)
	}
}

func TestSummarize_ProviderError(t *testing.T) {
	srv := newServer(t, &stubProvider{err: fmt.Errorf("timeout")})
	defer srv.Close()

	resp := postSummarize(t, srv, `{"text":"Texto valido"}`)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("esperava 502, obteve %d", resp.StatusCode)
	}
}

func TestHealth(t *testing.T) {
	srv := newServer(t, &stubProvider{})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("request falhou: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", resp.StatusCode)
	}
}
