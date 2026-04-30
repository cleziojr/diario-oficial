# Backend (API)

Servidor HTTP em Go (Chi + pgx + sqlc).

## API REST — documentos (metadados)

Prefixo: `/api/v1/documents` (rotas públicas por enquanto; autenticação JWT pode ser adicionada em outra tarefa).

| Método | Caminho | Descrição |
|--------|---------|-----------|
| `POST` | `/api/v1/documents` | Cria registro. Corpo: `{"filename":"..."}` |
| `GET` | `/api/v1/documents` | Lista paginada. Query: `page` (default 1), `limit` (default 20, máx. 100) |
| `GET` | `/api/v1/documents/{id}` | Detalhe por UUID |
| `DELETE` | `/api/v1/documents/{id}` | Remove por UUID; `404` se não existir |

Respostas JSON usam `id` (UUID string), `filename`, `created_at` (RFC3339Nano em UTC).

Listagem:

```json
{
  "items": [{"id":"...", "filename":"...", "created_at":"..."}],
  "page": 1,
  "limit": 20
}
```

## Outras rotas

- `GET /health` — liveness
- `GET /ready` — readiness (ping no Postgres)

## Desenvolvimento

- Gerar código sqlc: na raiz do repositório, `make sqlc` (requer Docker).
- Testes: `make backend-test` ou `cd backend && go test ./...`

Variáveis de ambiente: ver `pkg/config/config.go` (inclui `DATABASE_URL`, endereço HTTP).
