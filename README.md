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

## Rodando a stack completa

### Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) >= 24
- [Docker Compose](https://docs.docker.com/compose/) >= 2.20 (já incluso no Docker Desktop)

### Subir tudo com um comando

```bash
docker compose up --build
```

Aguarde o healthcheck do Postgres passar (~10s na primeira vez). A API sobe na porta **8080**.

### Verificar que está tudo ok

```bash
curl http://localhost:8080/ready
# {"status":"ready"}
```

### Variáveis de ambiente opcionais

Crie um arquivo `.env` na raiz (baseado em `.env.example`):

```env
POSTGRES_USER=diario
POSTGRES_PASSWORD=diario
POSTGRES_DB=diario_oficial
POSTGRES_PORT=5432
HTTP_PORT=8080
```

### Parar a stack

```bash
docker compose down          # mantém o volume de dados
docker compose down -v       # apaga também o volume do Postgres
```

## Desenvolvimento local (sem Docker)

```bash
# 1. Subir apenas o Postgres
docker compose up -d postgres

# 2. Configurar variáveis
cp backend/.env.example backend/.env

# 3. Rodar o servidor
cd backend && go run ./cmd/server

# 4. Testes
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
