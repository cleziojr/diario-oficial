export const TAG_LABELS: Record<string, string> = {
  licitacao: 'Licitação',
  contrato: 'Contrato',
  nomeacao: 'Nomeação',
  exoneracao: 'Exoneração',
  concurso: 'Concurso',
  decreto: 'Decreto',
  portaria: 'Portaria',
  lei: 'Lei',
  orcamento: 'Orçamento',
  convenio: 'Convênio',
  aposentadoria: 'Aposentadoria',
  obras: 'Obras',
  saude: 'Saúde',
  educacao: 'Educação',
  meio_ambiente: 'Meio Ambiente',
  outros: 'Outros',
}

export function tagLabel(slug: string): string {
  return TAG_LABELS[slug] ?? slug
}
