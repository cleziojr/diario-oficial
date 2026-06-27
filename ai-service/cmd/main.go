package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-service/internal/api"
	"ai-service/internal/provider"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("aviso: .env nao encontrado; usando variaveis do sistema")
	}

	p, err := provider.Load()
	if err != nil {
		log.Fatalf("erro ao carregar provider: %v", err)
	}
	log.Printf("provider ativo: %s", p.Name())

	addr := os.Getenv("AI_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(p),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("ai-service escutando em %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
