package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
		return &loggingProvider{inner: &openRouterProvider{}}, nil
	case "ollama":
		return &loggingProvider{inner: &ollamaProvider{}}, nil
	default:
		return nil, fmt.Errorf("provider desconhecido: %s", p)
	}
}

// loggingProvider wraps any LLMProvider and logs metadata for every call.
type loggingProvider struct {
	inner LLMProvider
}

func (l *loggingProvider) Name() string {
	return l.inner.Name()
}

func (l *loggingProvider) Summarize(ctx context.Context, text string) (string, error) {
	start := time.Now()
	inputChars := len([]rune(text))

	log.Printf("[llm] provider=%s input_chars=%d status=started", l.inner.Name(), inputChars)

	result, err := l.inner.Summarize(ctx, text)

	duration := time.Since(start)
	latencyMs := duration.Milliseconds()

	if err != nil {
		log.Printf("[llm] provider=%s input_chars=%d status=error latency_ms=%d error=%v",
			l.inner.Name(), inputChars, latencyMs, err)
		return "", err
	}

	outputChars := len([]rune(result))
	log.Printf("[llm] provider=%s input_chars=%d output_chars=%d status=ok latency_ms=%d",
		l.inner.Name(), inputChars, outputChars, latencyMs)

	return result, nil
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