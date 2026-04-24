# pdf-extraction

Módulo de extração de texto de PDFs do projeto [Insight Diário](https://github.com/cleziojr/diario-oficial). Recebe um arquivo PDF, extrai o texto de todas as páginas e salva o resultado em um arquivo `.txt`.

## Estrutura

```
pdf-extraction/
├── cmd/
│   └── main.go                        # Ponto de entrada — recebe o caminho do PDF via argumento
├── internal/
│   ├── extractor/
│   │   └── extractor.go               # Lógica de extração com github.com/dslipak/pdf
│   └── model/
│       └── extraction_result.go       # Struct com metadados da extração
├── tests/
│   └── extractor_test.go              # Testes unitários
├── input/                             # PDFs de entrada (não versionado)
├── output/                            # Arquivos .txt gerados (não versionado)
├── go.mod
└── README.md
```

## Pré-requisitos

- [Go 1.26+](https://go.dev/dl/)

## Instalando dependências

```bash
go mod tidy
```

## Executando

```bash
go run cmd/main.go <caminho-do-pdf>
```

Exemplo:

```bash
go run cmd/main.go input/diario-oficial.pdf
```

O arquivo `.txt` será gerado em `output/diario-oficial.txt`.

## Testando

```bash
go test ./tests/...
```

## Como funciona

1. O `main.go` recebe o caminho do PDF como argumento de linha de comando
2. `extractor.ExtractFromFile()` abre o PDF com a biblioteca `dslipak/pdf`
3. O texto é extraído página por página, com um cabeçalho `=== Página N ===` separando cada uma
4. O resultado é salvo em `output/<nome-original>.txt`
5. Metadados da extração (páginas, caracteres) são exibidos no terminal

## Biblioteca utilizada

[`github.com/dslipak/pdf`](https://pkg.go.dev/github.com/dslipak/pdf) — biblioteca pura Go para leitura de PDFs, sem dependências externas. Funciona bem para PDFs com texto digital (não escaneados). Para PDFs escaneados (imagens), seria necessário adicionar OCR.

## Limitações

- Não suporta PDFs protegidos por senha
- PDFs escaneados (apenas imagens) retornarão texto vazio — nesses casos é necessário OCR
- A ordem do texto extraído pode variar em PDFs com layout complexo (múltiplas colunas)

## Parte do projeto

Este módulo é responsável pela etapa de **extração de texto** do MVP do Insight Diário, fornecendo o conteúdo textual que será enviado ao `ai-service` para sumarização.
