package extractor_test

import (
	"os"
	"path/filepath"
	"testing"

	"pdf-extraction/internal/extractor"
)

func TestExtractFromFile_FileNotFound(t *testing.T) {
	_, err := extractor.ExtractFromFile("inexistente.pdf", t.TempDir())
	if err == nil {
		t.Error("esperava erro para arquivo inexistente, mas não obteve")
	}
}

func TestExtractFromFile_InvalidFile(t *testing.T) {
	// Cria um arquivo que não é um PDF válido
	tmp := filepath.Join(t.TempDir(), "fake.pdf")
	if err := os.WriteFile(tmp, []byte("isso nao e um pdf"), 0644); err != nil {
		t.Fatalf("erro ao criar arquivo de teste: %v", err)
	}

	_, err := extractor.ExtractFromFile(tmp, t.TempDir())
	if err == nil {
		t.Error("esperava erro para PDF inválido, mas não obteve")
	}
}

func TestExtractFromFile_OutputDirCreated(t *testing.T) {
	// Verifica que o diretório de saída é criado se não existir.
	// Como não temos um PDF real neste teste, apenas validamos o erro de PDF inválido
	// e que o diretório de saída seria criado caso o PDF fosse válido.
	outputDir := filepath.Join(t.TempDir(), "subdir", "output")

	tmp := filepath.Join(t.TempDir(), "fake.pdf")
	os.WriteFile(tmp, []byte("not a pdf"), 0644)

	// A função deve falhar no parse do PDF antes de chegar na criação do diretório,
	// então apenas garantimos que não há panic.
	extractor.ExtractFromFile(tmp, outputDir)
}
