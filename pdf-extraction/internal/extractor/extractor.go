package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pdf-extraction/internal/model"

	"github.com/dslipak/pdf"
)

// ExtractFromFile lê o PDF no caminho informado, extrai o texto de todas as
// páginas e salva o resultado em outputDir com o mesmo nome base e extensão .txt.
// Retorna um ExtractionResult com metadados da operação.
func ExtractFromFile(pdfPath string, outputDir string) (*model.ExtractionResult, error) {
	// Abre o PDF
	r, err := pdf.Open(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir PDF %q: %w", pdfPath, err)
	}

	totalPages := r.NumPage()
	if totalPages == 0 {
		return nil, fmt.Errorf("PDF %q não contém páginas", pdfPath)
	}

	// Extrai texto de cada página
	var sb strings.Builder
	for i := 1; i <= totalPages; i++ {
		page := r.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return nil, fmt.Errorf("erro ao extrair texto da página %d: %w", i, err)
		}

		sb.WriteString(fmt.Sprintf("=== Página %d ===\n", i))
		sb.WriteString(text)
		sb.WriteString("\n\n")
	}

	extracted := sb.String()

	// Garante que o diretório de saída existe
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de saída %q: %w", outputDir, err)
	}

	// Monta o caminho do arquivo de saída
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
	outputPath := filepath.Join(outputDir, baseName+".txt")

	// Salva o arquivo .txt
	if err := os.WriteFile(outputPath, []byte(extracted), 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo de saída %q: %w", outputPath, err)
	}

	return &model.ExtractionResult{
		SourceFile: pdfPath,
		OutputFile: outputPath,
		PageCount:  totalPages,
		CharCount:  len(extracted),
	}, nil
}
