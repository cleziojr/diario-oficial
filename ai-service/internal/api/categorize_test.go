package api

import "testing"

func TestParseMaterias_CleanJSON(t *testing.T) {
	raw := `{"materias":[{"title":"Portaria 1","summary":"resumo","category":"nomeacao","tags":["nomeacao"],"monetary_values":["R$ 10,00"],"dates":["2026-01-01"]}]}`
	ms, err := parseMaterias(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(ms) != 1 || ms[0].Title != "Portaria 1" || ms[0].Category != "nomeacao" {
		t.Fatalf("parse incorreto: %+v", ms)
	}
}

func TestParseMaterias_FencedAndNoise(t *testing.T) {
	raw := "Claro! Aqui esta:\n```json\n{\"materias\": [ {\"title\":\"A\",\"summary\":\"s\",\"category\":\"outros\",\"tags\":[]} ]}\n```\nFim."
	ms, err := parseMaterias(raw)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(ms) != 1 || ms[0].Title != "A" {
		t.Fatalf("parse incorreto: %+v", ms)
	}
}

func TestParseMaterias_Empty(t *testing.T) {
	ms, err := parseMaterias(`{"materias":[]}`)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(ms) != 0 {
		t.Fatalf("esperava 0 matérias, veio %d", len(ms))
	}
}

func TestParseMaterias_NoJSON(t *testing.T) {
	if _, err := parseMaterias("desculpe, nao posso ajudar"); err == nil {
		t.Fatal("esperava erro quando não há JSON")
	}
}
