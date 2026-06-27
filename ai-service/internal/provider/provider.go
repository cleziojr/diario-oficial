package provider

import (
	"context"
	"fmt"
	"os"

	"ai-service/internal/llm"
)

type LLMProvider interface {
	Summarize(ctx context.Context, text string) (string, error)
	Name() string
}

func Load() (LLMProvider, error) {
	p := os.Getenv("LLM_PROVIDER")
	switch p {
	case "openrouter", "":
		return &openRouterProvider{}, nil
	case "ollama":
		return &ollamaProvider{}, nil
	default:
		return nil, fmt.Errorf("provider desconhecido: %s", p)
	}
}

type openRouterProvider struct{}

func (o *openRouterProvider) Summarize(ctx context.Context, text string) (string, error) {
	return llm.Summarize(ctx, text)
}

func (o *openRouterProvider) Name() string {
	return "openrouter"
}

type ollamaProvider struct{}

func (o *ollamaProvider) Summarize(ctx context.Context, text string) (string, error) {
	return llm.SummarizeOllama(ctx, text)
}

func (o *ollamaProvider) Name() string {
	return "ollama"
}