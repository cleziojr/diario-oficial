package main

import (
	"fmt"
	"log"
	"os"

	"ai-service/internal/llm"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: não foi possível carregar .env")
	}

	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	if apiKey == "" {
		log.Fatal("HUGGINGFACE_API_KEY não definida")
	}

	fmt.Println("Chave carregada com sucesso!")

	text := getSampleText()

	summary, err := llm.Summarize(text)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n📄 TEXTO ORIGINAL:\n")
	fmt.Println(text)

	fmt.Println("\n🧠 RESUMO GERADO:\n")
	fmt.Println(summary)
}

func getSampleText() string {
	return `O governo do estado anunciou nesta terça-feira um novo pacote de medidas voltadas para a melhoria da infraestrutura urbana. 
O plano inclui investimentos em mobilidade, saneamento básico e modernização de serviços públicos digitais. 

Segundo o secretário de planejamento, as ações fazem parte de uma estratégia de longo prazo que busca aumentar a eficiência dos serviços prestados à população. 
Especialistas apontam que, embora as medidas sejam positivas, será necessário garantir a execução adequada dos projetos para que os resultados sejam efetivos.

A previsão é de que as primeiras obras sejam iniciadas ainda no próximo semestre.`
}