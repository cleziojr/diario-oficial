# Diário Oficial

Plataforma para extração, análise e consulta de publicações do Diário Oficial, com backend em Go, banco PostgreSQL e serviços de IA para sumarização.

## Stack

| Componente | Tecnologia |
|---|---|
| API | Go 1.23 + Chi + pgx |
| Banco de dados | PostgreSQL 16 |
| Extração de PDF | Go (pdf-extraction module) |
| Sumarização IA | Go + LLM (ai-service module) |
| Frontend | React + TypeScript + Vite |

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) >= 24
- [Docker Compose](https://docs.docker.com/compose/) >= 2.20 (já incluso no Docker Desktop)
- Mínimo **4 GB de RAM** disponível para o Docker (8 GB recomendado se usar Ollama)
- Mínimo **5 GB de espaço em disco** livre (o modelo llama3.2 ocupa ~2 GB)

## Configuração

Antes de subir a stack, crie o arquivo `.env` na raiz do projeto:

```bash
cp .env.example .env
```

Edite o `.env` conforme seu ambiente:

```env
# Banco de dados
POSTGRES_USER=diario
POSTGRES_PASSWORD=diario
POSTGRES_DB=diario_oficial
POSTGRES_PORT=5432

# API
HTTP_PORT=8080

# IA — escolha um provider: ollama (local) ou openrouter (nuvem)
LLM_PROVIDER=ollama

# Se LLM_PROVIDER=ollama
OLLAMA_MODEL=llama3.2

# Se LLM_PROVIDER=openrouter
OPENROUTER_API_KEY=
```

### Escolhendo o provider de IA

**Ollama (padrão — roda localmente, sem custo)**
- Não requer conta ou chave de API
- Na primeira execução, o modelo `llama3.2` (~2 GB) será baixado automaticamente — isso pode levar vários minutos dependendo da sua conexão
- Requer mais memória RAM — recomendado 8 GB disponíveis para o Docker

**OpenRouter (nuvem)**
- Requer uma conta em [openrouter.ai](https://openrouter.ai) e uma chave de API
- Mais rápido e leve para a máquina local
- Configure `LLM_PROVIDER=openrouter` e `OPENROUTER_API_KEY=sua-chave` no `.env`

## Rodando a stack completa

```bash
docker compose up --build
```

Na **primeira execução** com Ollama, aguarde a mensagem `model ready` nos logs antes de usar a API — o download do modelo leva alguns minutos. Você pode acompanhar com:

```bash
docker logs -f diario-oficial-ollama-pull
```

A API estará disponível na porta **8080** quando todos os containers estiverem saudáveis.

### Verificar que está tudo ok

```bash
curl http://localhost:8080/health
# {"status":"ok"}

curl http://localhost:8080/ready
# {"status":"ready"}
```

### Parar a stack

```bash
docker compose down        # mantém os dados (Postgres + modelo Ollama)
docker compose down -v     # apaga tudo, incluindo banco de dados e modelo baixado
```

> ⚠️ `docker compose down -v` apaga o banco de dados e o modelo Ollama. Na próxima execução, o modelo será baixado novamente.

## Desenvolvimento local (sem Docker)

```bash
# 1. Subir apenas o Postgres
docker compose up -d postgres

# 2. Configurar variáveis do backend
cp backend/.env.example backend/.env

# 3. Rodar o servidor
cd backend && go run ./cmd/server

# 4. Rodar os testes
make backend-test
```

## Endpoints principais

| Método | Caminho | Descrição |
|--------|---------|-----------|
| `GET` | `/health` | Liveness |
| `GET` | `/ready` | Readiness (ping no Postgres) |
| `GET` | `/api/v1/documents` | Lista documentos |
| `POST` | `/api/v1/documents` | Cria documento |
| `GET` | `/api/v1/documents/{id}` | Detalhe de documento |
| `PATCH` | `/api/v1/documents/{id}` | Atualiza documento |
| `DELETE` | `/api/v1/documents/{id}` | Remove documento |
| `GET` | `/api/v1/documents/{id}/insights` | Insights agregados do documento |
| `POST` | `/api/v1/documents/{id}/analyses` | Cria análise (dispara sumarização via IA) |
| `GET` | `/api/v1/documents/{id}/analyses` | Lista análises do documento |
| `GET` | `/api/v1/analyses/{id}` | Detalhe de análise |
| `PATCH` | `/api/v1/analyses/{id}` | Atualiza análise |
| `DELETE` | `/api/v1/analyses/{id}` | Remove análise |

Veja a documentação completa em [backend/README.md](backend/README.md).

## Módulos do repositório

| Módulo | Descrição |
|---|---|
| [backend/](backend/) | API REST (Chi + pgx + sqlc) |
| [pdf-extraction/](pdf-extraction/) | Extração de texto de PDFs |
| [ai-service/](ai-service/) | Sumarização via LLM |
| [frontend/](frontend/) | Interface web (React + Vite) |

## CI/CD e DevOps

- Pipeline: [.github/workflows/ci.yml](.github/workflows/ci.yml)
- Métricas DORA: [docs/devops/DORA.md](docs/devops/DORA.md)
- Retrospectiva Sprint 02: [docs/sprints/SPRINT-02-RETRO.md](docs/sprints/SPRINT-02-RETRO.md)
- **Acompanhamento do projeto:** [GitHub Project](https://github.com/users/cleziojr/projects/1)