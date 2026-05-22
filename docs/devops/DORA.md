# Métricas DORA — Diário Oficial

As quatro métricas DORA (_DevOps Research and Assessment_) medem a performance de entrega de software. Este documento define como cada métrica é coletada e interpretada neste projeto.

## Métricas implementadas

### 1. Frequência de Deploy (Deployment Frequency)

**Definição:** Com que frequência o time faz deploy em produção.

**Como medimos:** Cada push na branch `main` que passa pelo pipeline CI completo (`backend-test` + `backend-lint` + `docker-smoke`) equivale a um deploy. O job `record-deploy` registra o evento no GitHub Actions Summary.

**Referência DORA:**

| Nível | Frequência |
|---|---|
| Elite | Múltiplas vezes por dia |
| Alto | Entre uma vez por semana e uma vez por dia |
| Médio | Entre uma vez por mês e uma vez por semana |
| Baixo | Menos de uma vez por mês |

**Meta para Sprint 02:** pelo menos 1 deploy por semana (nível Alto).

**Como consultar:** Acesse a aba **Actions → CI → record-deploy → Summary** de qualquer run em `main` para ver o contador do sprint.

---

### 2. Lead Time for Changes

**Definição:** Tempo médio desde o commit até o código estar em produção.

**Como medimos:** O step `Compute lead time` no job `record-deploy` calcula a diferença entre o timestamp do último commit em `main` e o momento do deploy (execução do job). O valor é exibido em minutos no GitHub Step Summary.

**Fórmula:**

```
lead_time = timestamp_deploy − timestamp_merge_commit  (em minutos)
```

**Referência DORA:**

| Nível | Lead Time |
|---|---|
| Elite | < 1 hora |
| Alto | Entre 1 dia e 1 semana |
| Médio | Entre 1 semana e 1 mês |
| Baixo | > 1 mês |

**Meta para Sprint 02:** lead time < 60 minutos por PR (nível Elite).

**Como consultar:** GitHub Actions → run em `main` → job `record-deploy` → Step Summary.

---

## Métricas futuras (backlog)

| Métrica | Descrição | Status |
|---|---|---|
| Change Failure Rate | % de deploys que causam incidentes | Backlog |
| Time to Restore Service | Tempo médio para recuperar de uma falha | Backlog |

---

## Histórico por Sprint

| Sprint | Deploys | Lead Time médio |
|---|---|---|
| Sprint 01 | — | — |
| Sprint 02 | Em andamento | Em andamento |

> **Nota:** As métricas são coletadas automaticamente a cada push em `main`. Atualize a tabela ao final de cada sprint com os valores consolidados do GitHub Actions.
