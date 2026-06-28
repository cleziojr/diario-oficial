package prompt

import "fmt"

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