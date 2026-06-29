package taxonomy

import (
	"reflect"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"licitacao":   "licitacao",
		"  Licitacao": "licitacao",
		"MEIO AMBIENTE": "meio_ambiente",
		"qualquer-coisa": "outros",
		"":              "outros",
	}
	for in, want := range cases {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeListDedupAndFallback(t *testing.T) {
	got := NormalizeList([]string{"licitacao", "Licitacao", "xpto"})
	want := []string{"licitacao", "outros"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizeList = %v, want %v", got, want)
	}
	if got := NormalizeList(nil); !reflect.DeepEqual(got, []string{"outros"}) {
		t.Errorf("NormalizeList(nil) = %v, want [outros]", got)
	}
}

func TestLabel(t *testing.T) {
	if Label("saude") != "Saúde" {
		t.Errorf("Label(saude) = %q", Label("saude"))
	}
	if Label("inexistente") != "Outros" {
		t.Errorf("Label fallback = %q", Label("inexistente"))
	}
}
