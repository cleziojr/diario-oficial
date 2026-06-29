// Package taxonomy define o vocabulário controlado de tags/categorias usado
// para classificar matérias de Diários Oficiais municipais.
package taxonomy

import "strings"

// labels mapeia cada slug válido para seu rótulo PT-BR.
var labels = map[string]string{
	"licitacao":     "Licitação",
	"contrato":      "Contrato",
	"nomeacao":      "Nomeação",
	"exoneracao":    "Exoneração",
	"concurso":      "Concurso",
	"decreto":       "Decreto",
	"portaria":      "Portaria",
	"lei":           "Lei",
	"orcamento":     "Orçamento",
	"convenio":      "Convênio",
	"aposentadoria": "Aposentadoria",
	"obras":         "Obras",
	"saude":         "Saúde",
	"educacao":      "Educação",
	"meio_ambiente": "Meio Ambiente",
	"outros":        "Outros",
}

// ordem estável das tags (para listagens determinísticas).
var order = []string{
	"licitacao", "contrato", "nomeacao", "exoneracao", "concurso", "decreto",
	"portaria", "lei", "orcamento", "convenio", "aposentadoria", "obras",
	"saude", "educacao", "meio_ambiente", "outros",
}

// Fallback é a categoria usada quando o modelo devolve algo fora do vocabulário.
const Fallback = "outros"

// Label retorna o rótulo PT-BR de um slug; para slugs desconhecidos devolve o
// rótulo de "outros".
func Label(slug string) string {
	if l, ok := labels[slug]; ok {
		return l
	}
	return labels[Fallback]
}

// IsValid informa se o slug pertence ao vocabulário.
func IsValid(slug string) bool {
	_, ok := labels[slug]
	return ok
}

// Normalize converte uma tag arbitrária para um slug válido, ou Fallback.
func Normalize(tag string) string {
	t := strings.ToLower(strings.TrimSpace(tag))
	t = strings.ReplaceAll(t, " ", "_")
	if IsValid(t) {
		return t
	}
	return Fallback
}

// NormalizeList normaliza, deduplica e descarta vazios de uma lista de tags.
// Garante ao menos [Fallback] quando a entrada não produz nenhuma tag válida.
func NormalizeList(tags []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		n := Normalize(t)
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return []string{Fallback}
	}
	return out
}

// All devolve os slugs na ordem canônica (útil para a UI montar filtros).
func All() []string {
	cp := make([]string, len(order))
	copy(cp, order)
	return cp
}
