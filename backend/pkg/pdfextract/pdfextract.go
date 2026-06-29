// Package pdfextract extrai texto de PDFs usando o poppler (pdftotext/pdfinfo),
// disponível no runtime do backend.
package pdfextract

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Result é o texto extraído de um PDF e a contagem de páginas.
type Result struct {
	Text  string
	Pages int
}

// Extract grava os bytes em um arquivo temporário e roda pdftotext/pdfinfo.
func Extract(ctx context.Context, data []byte) (Result, error) {
	if len(data) == 0 {
		return Result{}, fmt.Errorf("pdf vazio")
	}

	f, err := os.CreateTemp("", "ingest-*.pdf")
	if err != nil {
		return Result{}, fmt.Errorf("criar temp: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(data); err != nil {
		f.Close()
		return Result{}, fmt.Errorf("gravar temp: %w", err)
	}
	if err := f.Close(); err != nil {
		return Result{}, fmt.Errorf("fechar temp: %w", err)
	}

	// Texto: pdftotext <arquivo> - (stdout)
	var out, errBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, "pdftotext", "-enc", "UTF-8", f.Name(), "-")
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("pdftotext: %w (%s)", err, strings.TrimSpace(errBuf.String()))
	}

	pages := countPages(ctx, f.Name())

	return Result{Text: out.String(), Pages: pages}, nil
}

// countPages usa pdfinfo; em caso de falha devolve 0 (não é fatal).
func countPages(ctx context.Context, path string) int {
	var out bytes.Buffer
	cmd := exec.CommandContext(ctx, "pdfinfo", path)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0
	}
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			n, _ := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
			return n
		}
	}
	return 0
}
