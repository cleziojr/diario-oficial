package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"ai-service/internal/provider"
)

type summarizeRequest struct {
	Text string `json:"text"`
}

type summarizeResponse struct {
	Summary string `json:"summary"`
	Model   string `json:"model"`
}

type summarizeHandler struct {
	provider provider.LLMProvider
}

func (h *summarizeHandler) handle(w http.ResponseWriter, r *http.Request) {
	var body summarizeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON invalido")
		return
	}

	body.Text = strings.TrimSpace(body.Text)
	if body.Text == "" {
		writeError(w, http.StatusBadRequest, "campo text e obrigatorio")
		return
	}

	summary, err := h.provider.Summarize(r.Context(), body.Text)
	if err != nil {
		writeError(w, http.StatusBadGateway, "erro ao gerar resumo: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, summarizeResponse{
		Summary: summary,
		Model:   h.provider.Name(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
