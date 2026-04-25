package model

// ExtractionResult representa o resultado da extração de um PDF.
type ExtractionResult struct {
	SourceFile string
	OutputFile string
	PageCount  int
	CharCount  int
}
