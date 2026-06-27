package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"ai-service/internal/llm"

	"github.com/joho/godotenv"
)

const maxRetries = 3

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: .env nao encontrado; usando variaveis do sistema")
	}

	text := getSampleText()
	ctx := context.Background()

	var (
		summary string
		err     error
	)
	for i := 1; i <= maxRetries; i++ {
		summary, err = llm.Summarize(ctx, text)
		if err == nil {
			break
		}
		log.Printf("tentativa %d/%d falhou: %v", i, maxRetries, err)
		if i < maxRetries {
			time.Sleep(time.Duration(i) * 2 * time.Second)
		}
	}
	if err != nil {
		log.Fatalf("erro apos %d tentativas: %v", maxRetries, err)
	}

	fmt.Println("\nRESUMO GERADO:")
	fmt.Println(summary)
}

func getSampleText() string {
	if v := os.Getenv("INPUT_TEXT"); v != "" {
		return v
	}
	return `O governo do estado anunciou nesta terca-feira um novo pacote de medidas voltadas para a melhoria da infraestrutura urbana.
O plano inclui investimentos em mobilidade, saneamento basico e modernizacao de servicos publicos digitais.`
}