package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"pdf-extraction/internal/backend"
	"pdf-extraction/internal/extractor"
)

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

	c := backend.New(backendURL)
	if err := c.PostAnalysis(context.Background(), documentID, string(extracted)); err != nil {
		log.Fatalf("erro ao enviar ao backend: %v", err)
	}

	fmt.Println("Analise enviada ao backend com sucesso.")
}