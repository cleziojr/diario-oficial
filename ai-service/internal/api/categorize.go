package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"ai-service/internal/llm"
)

type categorizeRequest struct {
	Text string `json:"text"`
	Page int    `json:"page"`
}

type entities struct {
	People []string `json:"people"`
	Orgs   []string `json:"orgs"`
}

type materia struct {
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags"`
	Entities       entities `json:"entities"`
	MonetaryValues []string `json:"monetary_values"`
	Dates          []string `json:"dates"`
}

type categorizeResponse struct {
	Model    string    `json:"model"`
	Materias []materia `json:"materias"`
}

type categorizeHandler struct{}

func (h *categorizeHandler) handle(w http.ResponseWriter, r *http.Request) {
	var body categorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "corpo JSON invalido")
		return
	}
	body.Text = strings.TrimSpace(body.Text)
	if body.Text == "" {
		writeError(w, http.StatusBadRequest, "campo text e obrigatorio")
		return
	}

	provider := os.Getenv("LLM_PROVIDER")
	var raw string
	var err error
	switch provider {
	case "ollama":
		raw, err = llm.CategorizeOllama(r.Context(), body.Text)
	default: // openrouter (padrao)
		raw, err = llm.CategorizeOpenRouter(r.Context(), body.Text)
		provider = "openrouter"
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "erro ao categorizar: "+err.Error())
		return
	}

	materias, err := parseMaterias(raw)
	if err != nil {
		writeError(w, http.StatusBadGateway, "resposta do modelo nao e JSON valido: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, categorizeResponse{Model: provider, Materias: materias})
}

// parseMaterias extrai o objeto JSON da resposta do modelo (tolerante a cercas de
// código e texto ao redor) e devolve a lista de matérias.
func parseMaterias(raw string) ([]materia, error) {
	jsonStr := extractJSONObject(raw)
	if jsonStr == "" {
		return nil, fmt.Errorf("nenhum objeto JSON encontrado")
	}
	var parsed struct {
		Materias []materia `json:"materias"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, err
	}
	if parsed.Materias == nil {
		parsed.Materias = []materia{}
	}
	return parsed.Materias, nil
}

// extractJSONObject devolve o primeiro objeto JSON balanceado dentro de s.
func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	inStr := false
	esc := false
	for i := start; i < len(s); i++ {
		c := s[i]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
