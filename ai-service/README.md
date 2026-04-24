# ai-service

Módulo de sumarização com LLM. Recebe um texto extraído de Diários Oficiais e retorna um resumo em tópicos gerado por IA.

## Estrutura

```
ai-service/
├── cmd/
│   └── main.go                  # Ponto de entrada da aplicação
├── internal/
│   ├── llm/
│   │   └── llm_service.go       # Integração com a API do OpenRouter
│   ├── model/
│   │   └── summary.go           # Struct de resposta
│   └── prompt/
│       └── prompt_builder.go    # Construção de prompts
├── tests/
│   └── llm_test.go              # Testes unitários com mock HTTP
├── go.mod
└── go.sum
```

## Pré-requisitos

- [Go 1.26+](https://go.dev/dl/)
- Conta no [OpenRouter](https://openrouter.ai) com uma API key

## Configuração

Crie um arquivo `.env` na raiz do módulo:

```env
OPENROUTER_API_KEY=sk-or-...
```

> A chave pode ser gerada em [openrouter.ai/keys](https://openrouter.ai/keys). Não é necessário adicionar créditos para usar modelos gratuitos.

## Executando

```bash
go run cmd/main.go
```

## Testando

```bash
go test ./tests/...
```

Os testes usam um servidor HTTP mock e não fazem chamadas reais à API.

## Como funciona

1. O `main.go` carrega a chave via `.env` e passa um texto de exemplo para `llm.Summarize()`
2. `llm_service.go` monta uma requisição no formato OpenAI-compatible e envia para `https://openrouter.ai/api/v1/chat/completions`
3. O OpenRouter roteia para um modelo disponível e retorna o resumo em tópicos

## Modelo utilizado

O serviço usa `openrouter/auto`, que seleciona automaticamente um modelo disponível no momento. Para garantir custo zero, substitua por um modelo gratuito explícito em `internal/llm/llm_service.go`:

```go
Model: "meta-llama/llama-3.3-70b-instruct:free",
```

Modelos gratuitos disponíveis: [openrouter.ai/models?q=free](https://openrouter.ai/models?q=free)

## Dependências

| Pacote | Uso |
|---|---|
| `github.com/joho/godotenv` | Carregamento do `.env` |

## Parte do projeto

Este módulo integra o MVP do Insight Diário, responsável pela etapa de **resumo com LLM** após a extração de texto dos PDFs de Diários Oficiais.
