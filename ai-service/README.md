# ai-service

Módulo de sumarização com LLM. Recebe um texto extraído de Diários Oficiais e retorna um resumo em tópicos gerado por IA, exposto via HTTP para integração com o backend.

## Estrutura

```
ai-service/
├── cmd/
│   ├── main.go                      # CLI de teste (uso local)
│   └── server/
│       └── main.go                  # Servidor HTTP (produção)
├── internal/
│   ├── api/
│   │   ├── router.go                # Rotas: GET /health, POST /summarize
│   │   ├── summarize.go             # Handler POST /summarize
│   │   └── summarize_test.go        # Testes do handler
│   ├── llm/
│   │   └── llm_service.go           # Ponto de entrada do pacote llm
│   ├── model/
│   │   └── summary.go               # Struct de resposta
│   ├── prompt/
│   │   └── prompt_builder.go        # Construção centralizada de prompts
│   └── provider/
│       ├── provider.go              # Interface LLMProvider + loader
│       ├── openrouter.go            # Provider OpenRouter (retry + backoff)
│       ├── ollama.go                # Provider Ollama (local)
│       ├── http.go                  # Helper HTTP compartilhado
│       ├── prompt.go                # Prompt interno do provider
│       └── types.go                 # Tipos OpenAI Chat format
├── tests/
│   └── llm_test.go                  # Testes unitários com mock HTTP
├── Dockerfile
├── .env.example
├── go.mod
└── go.sum
```

## Pré-requisitos

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) para execução via Compose

Um dos seguintes providers:

- **Ollama** (padrão) — execução local, sem custo, sem dependência externa
- **OpenRouter** — requer conta em [openrouter.ai](https://openrouter.ai) com API key

## Configuração

Copie o arquivo de exemplo e preencha conforme o provider escolhido:

```bash
cp .env.example .env
```

### Ollama (padrão)

```env
LLM_PROVIDER=ollama
OLLAMA_HOST=http://ollama:11434
OLLAMA_MODEL=llama3.2
```

### OpenRouter

```env
LLM_PROVIDER=openrouter
OPENROUTER_API_KEY=sk-or-...
```

> A chave pode ser gerada em [openrouter.ai/keys](https://openrouter.ai/keys). Modelos gratuitos disponíveis em [openrouter.ai/models?q=free](https://openrouter.ai/models?q=free).

## Executando via Docker Compose

```bash
# Sobe postgres + api + ai-service + ollama
docker compose --profile ollama up -d

# Baixa o modelo (apenas na primeira execução, ~2 GB)
docker exec diario-oficial-ollama ollama pull llama3.2
```

## Executando localmente

```bash
# Servidor HTTP
go run cmd/server/main.go

# CLI de teste
go run cmd/main.go
```

## API

### `GET /health`

Verifica se o serviço está no ar.

```json
{ "status": "ok" }
```

### `POST /summarize`

Gera um resumo estruturado a partir de um trecho de Diário Oficial.

**Request:**
```json
{ "text": "O governo do estado anunciou..." }
```

**Response `200`:**
```json
{
  "summary": "- Ato administrativo: ...\n- Valor: R$ ...",
  "model": "ollama(llama3.2)"
}
```

**Response `400`:** campo `text` ausente ou vazio.

**Response `502`:** erro ao chamar o provider LLM.

## Testando

```bash
# Testes do handler (sem chamadas reais à API)
go test ./internal/api/...

# Testes do pacote llm
go test ./tests/...
```

## Como funciona

1. `cmd/server/main.go` carrega o provider via `LLM_PROVIDER` e sobe o servidor na porta `9090`
2. O backend chama `POST /summarize` com o texto extraído do PDF
3. O handler delega ao provider ativo (`ollama` ou `openrouter`)
4. O provider monta a requisição no formato OpenAI Chat e retorna o resumo
5. O resumo é devolvido ao backend para persistência em `document_analyses`

## Dependências

| Pacote | Uso |
|---|---|
| `github.com/go-chi/chi/v5` | Roteamento HTTP |
| `github.com/joho/godotenv` | Carregamento do `.env` |

## Parte do projeto

Este módulo integra o MVP do Insight Diário, responsável pela etapa de **resumo com LLM** após a extração de texto dos PDFs de Diários Oficiais. É chamado exclusivamente pelo backend via rede interna Docker — não é exposto para o frontend.