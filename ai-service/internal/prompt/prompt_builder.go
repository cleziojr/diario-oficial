package prompt

import "fmt"

// categories é o vocabulário controlado que o modelo DEVE usar ao classificar.
const categories = "licitacao, contrato, nomeacao, exoneracao, concurso, decreto, portaria, lei, orcamento, convenio, aposentadoria, obras, saude, educacao, meio_ambiente, outros"

// BuildCategorize monta o prompt que pede ao modelo a quebra do trecho do Diário
// Oficial em "matérias" categorizadas, devolvendo EXCLUSIVAMENTE um JSON.
func BuildCategorize(text string) string {
	return fmt.Sprintf(`Você é um assistente que analisa Diários Oficiais de municípios brasileiros.

Quebre o trecho abaixo em "matérias" (atos/publicações independentes) e classifique cada uma.

Responda EXCLUSIVAMENTE com um objeto JSON válido, sem texto antes ou depois, no formato:
{"materias":[{"title":"string","summary":"string","category":"<uma das categorias>","tags":["<categorias>"],"entities":{"people":["string"],"orgs":["string"]},"monetary_values":["R$ ..."],"dates":["YYYY-MM-DD"]}]}

Regras:
- "category" e "tags" DEVEM usar apenas estes slugs: %s
- Use "outros" quando nenhum slug se aplicar.
- "title": título curto e objetivo da matéria.
- "summary": resumo de 1-2 frases. Não copie o texto original.
- "monetary_values": valores em reais encontrados (lista vazia se não houver).
- "dates": datas no formato YYYY-MM-DD (lista vazia se não houver).
- Se o trecho não contiver matérias relevantes, devolva {"materias":[]}.

TEXTO:
%s`, categories, text)
}

func Build(text string) string {
	return fmt.Sprintf(`Você é um assistente especializado em análise de Diários Oficiais brasileiros.

Analise o trecho abaixo e produza um resumo estruturado em tópicos claros, destacando:
- Atos administrativos relevantes (nomeações, exonerações, contratos, licitações)
- Valores financeiros mencionados
- Prazos e datas importantes
- Órgãos e entidades envolvidos

Seja objetivo. Não repita o texto original.

TEXTO:
%s`, text)
}