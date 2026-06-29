package extractor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pdf-extraction/internal/model"
)

// ExtractFromFile lê o PDF no caminho informado, extrai o texto de todas as
// páginas usando pdftotext e salva o resultado em outputDir com o mesmo nome
// base e extensão .txt.
// Retorna um ExtractionResult com metadados da operação.
func ExtractFromFile(pdfPath string, outputDir string) (*model.ExtractionResult, error) {
	// Garante que o diretório de saída existe
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de saída %q: %w", outputDir, err)
	}

	// Monta o caminho do arquivo de saída
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	outputPath := filepath.Join(outputDir, baseName+".txt")

	fmt.Println("Executando pdftotext...")

	// Executa pdftotext
	cmd := exec.Command("pdftotext", "-layout", pdfPath, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar pdftotext: %w", err)
	}

	extracted, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo extraido: %w", err)
	}

	if len(extracted) == 0 {
		return nil, fmt.Errorf("nenhum texto extraído do PDF %q", pdfPath)
	}

	// Conta páginas via pdfinfo
	pageCount := 0
	out, err := exec.Command("pdfinfo", pdfPath).Output()
	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.HasPrefix(line, "Pages:") {
				fmt.Sscanf(strings.TrimPrefix(line, "Pages:"), "%d", &pageCount)
			}
		}
	}

	fmt.Printf("Total de páginas: %d\n", pageCount)

	return &model.ExtractionResult{
		SourceFile: pdfPath,
		OutputFile: outputPath,
		PageCount:  pageCount,
		CharCount:  len(extracted),
	}, nil
}