# Retrospectiva — Sprint 02

**Data:** 2026-05-22  
**Time:** Deriki Pereira, Clézio Jr.  
**Formato:** Start / Stop / Continue + Plano de ação

---

## Contexto do Sprint

Sprint 02 focou em:
- Containerização do backend com Docker + Compose
- Expansão do pipeline CI/CD (lint, smoke test, métricas DORA)
- Documentação de DevOps (DORA.md, README raiz)

---

## O que foi bem (Continue)

- **Velocidade de entrega:** os três módulos Go (backend, pdf-extraction, ai-service) já tinham estrutura clara, o que facilitou a criação do Dockerfile sem refatorações.
- **Healthcheck no Compose:** usar `depends_on: condition: service_healthy` eliminou race conditions entre API e Postgres desde o início.
- **Endpoint `/ready` desde o Sprint 01:** já existia e tornou o smoke test no CI trivial de implementar.

---

## O que atrapalhou (Stop)

- **Ausência de README raiz:** o repositório não tinha ponto de entrada para novos colaboradores; cada módulo tinha sua própria doc isolada.
- **Pipeline de CI sem lint:** PRs com problemas de estilo passavam sem bloqueio, gerando retrabalho no code review.
- **Métricas de qualidade não rastreadas:** sem DORA, não havia visibilidade sobre frequência de deploy nem lead time.

---

## O que iniciar (Start)

- Adicionar relatório de cobertura de testes ao CI e publicar badge no README.
- Definir workflow de PR com template (descrição, checklist de testes, link para issue).
- Criar ambiente de staging (compose profile separado) para validar antes de mergear em `main`.

---

## Plano de ação

| # | Ação | Responsável | Prazo | Status |
|---|---|---|---|---|
| 1 | Adicionar `coverage.out` ao CI e badge de cobertura no README raiz | Deriki | Sprint 03 | Pendente |
| 2 | Criar template de PR em `.github/pull_request_template.md` com checklist (testes, lint, docs) | Clézio | Sprint 03 | Pendente |
| 3 | Configurar compose profile `staging` para testar migrações novas sem afetar volume de dev | Deriki | Sprint 03 | Pendente |

---

## Roteiro do vídeo de demonstração

> Use este roteiro para gravar o vídeo de entrega do Sprint 02 (~3–5 min).

1. **(0:00)** Introdução: o que foi o Sprint 02 e qual o objetivo.
2. **(0:30)** Mostrar o `Dockerfile` do backend — explicar o multi-stage build (builder → alpine) e por que isso reduz o tamanho da imagem final.
3. **(1:00)** Executar `docker compose up --build` a partir da raiz e mostrar os logs subindo (Postgres healthcheck → API online).
4. **(1:45)** Rodar `curl http://localhost:8080/ready` e mostrar `{"status":"ready"}` no terminal.
5. **(2:00)** Abrir o GitHub Actions e mostrar o pipeline verde com os quatro jobs: `backend-test`, `backend-lint`, `docker-smoke`, `record-deploy`.
6. **(2:45)** Abrir o Step Summary do job `record-deploy` e mostrar as duas métricas DORA registradas automaticamente.
7. **(3:15)** Abrir `docs/devops/DORA.md` e explicar brevemente as métricas e metas.
8. **(3:45)** Abrir esta retrospectiva (`SPRINT-02-RETRO.md`) e ler o plano de ação — reforçar o que vem no Sprint 03.
9. **(4:15)** Mostrar o GitHub Project com as issues do Sprint 02 fechadas.
10. **(4:30)** Encerramento e próximos passos.
