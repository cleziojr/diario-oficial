package main

import (
	"fmt"
	"log"
	"os"

	"pdf-extraction/internal/extractor"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("uso: go run cmd/main.go <caminho-do-pdf>")
	}

	pdfPath := os.Args[1]
	outputDir := "output"

	fmt.Printf("📄 Extraindo texto de: %s\n", pdfPath)

	result, err := extractor.ExtractFromFile(pdfPath, outputDir)
	if err != nil {
		log.Fatalf("erro na extração: %v", err)
	}

	fmt.Printf("✅ Extração concluída!\n")
	fmt.Printf("   Páginas processadas : %d\n", result.PageCount)
	fmt.Printf("   Caracteres extraídos: %d\n", result.CharCount)
	fmt.Printf("   Arquivo gerado      : %s\n", result.OutputFile)
}
