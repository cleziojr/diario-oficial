# Backend (API)

Servidor HTTP em Go (Chi + pgx + sqlc).

## API REST — documentos (metadados)

Prefixo: `/api/v1/documents` (rotas públicas por enquanto; autenticação JWT pode ser adicionada em outra tarefa).

| Método | Caminho | Descrição |
|--------|---------|-----------|
| `POST` | `/api/v1/documents` | Cria registro. Corpo: `{"filename":"..."}` |
| `GET` | `/api/v1/documents` | Lista paginada. Query: `page` (default 1), `limit` (default 20, máx. 100) |
| `GET` | `/api/v1/documents/{id}` | Detalhe por UUID |
| `GET` | `/api/v1/documents/{id}/insights` | Retorna JSON aninhado com o documento e um array agregado de insights |
| `PATCH` | `/api/v1/documents/{id}` | Atualiza `filename`. Corpo: `{"filename":"..."}` |
| `DELETE` | `/api/v1/documents/{id}` | Remove por UUID; `404` se não existir |

Respostas JSON usam `id` (UUID string), `filename`, `created_at` (RFC3339Nano em UTC). O campo `updated_at` aparece apenas quando o documento foi atualizado via PATCH.

Listagem:

```json
{
  "items": [{"id":"...", "filename":"...", "created_at":"..."}],
  "page": 1,
  "limit": 20
}
```

Documento com insights agregados (útil para demonstrar no vídeo):

```bash
curl http://localhost:8080/api/v1/documents/{id}/insights
```

```json
{
  "document": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "filename": "diario.pdf",
    "created_at": "2026-05-22T10:00:00Z"
  },
  "insights": [
    {"type": "keyword", "value": "licitação"},
    {"type": "deadline", "value": "2026-06-01"}
  ]
}
```

O endpoint agrega os arrays `insights` de todas as análises do documento. Se o documento não existir, retorna `404`; se o UUID for inválido, retorna `400`.

## Outras rotas

- `GET /health` — liveness
- `GET /ready` — readiness (ping no Postgres)

## Desenvolvimento

- Gerar código sqlc: na raiz do repositório, `make sqlc` (requer Docker).
- Testes: `make backend-test` ou `cd backend && go test ./...`

Variáveis de ambiente: ver `pkg/config/config.go` (inclui `DATABASE_URL`, endereço HTTP).
