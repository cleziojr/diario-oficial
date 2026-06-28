package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"pdf-extraction/internal/backend"
	"pdf-extraction/internal/extractor"
)

const (
	chunkSize  = 3000
	chunkOverlap = 200
)

func splitIntoChunks(text string) []string {
	var chunks []string
	runes := []rune(text)
	total := len(runes)
	step := chunkSize - chunkOverlap

	for start := 0; start < total; start += step {
		end := start + chunkSize
		if end > total {
			end = total
		}
		chunks = append(chunks, string(runes[start:end]))
		if end == total {
			break
		}
	}
	return chunks
}

func main() {
	pdfPath := ""
	if len(os.Args) >= 2 {
		pdfPath = os.Args[1]
	} else {
		pdfPath = os.Getenv("PDF_PATH")
	}

	if pdfPath == "" {
		log.Fatal("uso: pdf-extraction <caminho-do-pdf> ou defina PDF_PATH no ambiente")
	}

	outputDir := "output"

	fmt.Printf("Extraindo texto de: %s\n", pdfPath)

	result, err := extractor.ExtractFromFile(pdfPath, outputDir)
	if err != nil {
		log.Fatalf("erro na extracao: %v", err)
	}

	fmt.Printf("Paginas : %d\n", result.PageCount)
	fmt.Printf("Chars   : %d\n", result.CharCount)
	fmt.Printf("Arquivo : %s\n", result.OutputFile)

	documentID := ""
	if len(os.Args) >= 3 {
		documentID = os.Args[2]
	} else {
		documentID = os.Getenv("DOCUMENT_ID")
	}

	if documentID == "" {
		log.Println("aviso: DOCUMENT_ID nao definido; texto nao enviado ao backend")
		return
	}

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://api:8080"
	}

	extracted, err := os.ReadFile(result.OutputFile)
	if err != nil {
		log.Fatalf("erro ao ler arquivo extraido: %v", err)
	}

	chunks := splitIntoChunks(string(extracted))
	fmt.Printf("Enviando %d chunk(s) ao backend...\n", len(chunks))

	c := backend.New(backendURL)
	var summaries []string

	for i, chunk := range chunks {
		fmt.Printf("[%d/%d] Enviando chunk ao backend...\n", i+1, len(chunks))
		summary, err := c.PostAnalysis(context.Background(), documentID, chunk)
		if err != nil {
			log.Fatalf("erro ao enviar chunk %d ao backend: %v", i+1, err)
		}
		summaries = append(summaries, summary)
		fmt.Printf("[%d/%d] Resumo parcial recebido.\n", i+1, len(chunks))
	}

	fmt.Printf("\nResumo consolidado (%d partes):\n\n%s\n", len(summaries), strings.Join(summaries, "\n\n---\n\n"))
	fmt.Println("Analise enviada ao backend com sucesso.")
}