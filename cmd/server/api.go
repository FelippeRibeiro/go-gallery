package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/FelippeRibeiro/go-gallery/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] Arquivo .env não encontrado, usando variáveis do ambiente do sistema")
	}

	// Sem JWT_SECRET o middleware cai num segredo de desenvolvimento conhecido,
	// que permitiria forjar tokens. Aceitável só em dev — em produção
	// (COOKIE_SECURE=true) é erro fatal.
	if os.Getenv("JWT_SECRET") == "" {
		secure := strings.ToLower(os.Getenv("COOKIE_SECURE"))
		if secure == "true" || secure == "1" {
			log.Fatal("JWT_SECRET é obrigatório em produção (COOKIE_SECURE=true)")
		}
		log.Println("[WARN] JWT_SECRET não definido — usando segredo de desenvolvimento INSEGURO")
	}

	ctx := context.Background()

	if err := db.Init(); err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	s3client.ObterS3Client()
	if err := s3client.EnsureBucket(ctx); err != nil {
		log.Fatalf("Erro ao garantir bucket S3: %v", err)
	}
	if err := s3client.EnsurePublicReadPolicy(ctx); err != nil {
		log.Fatalf("Erro ao garantir política de leitura pública: %v", err)
	}

	// ReadHeaderTimeout protege contra slowloris; IdleTimeout recicla conexões
	// keep-alive ociosas. Sem ReadTimeout/WriteTimeout globais de propósito:
	// matariam uploads lentos e as conexões WebSocket.
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           server.New(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Servidor iniciado em http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
