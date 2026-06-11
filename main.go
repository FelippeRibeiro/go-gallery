package main

import (
	"context"
	"log"
	"net/http"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/FelippeRibeiro/go-gallery/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] Arquivo .env não encontrado, usando variáveis do ambiente do sistema")
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

	srv := server.New()
	log.Println("Servidor iniciado em http://localhost:8080")
	if err := http.ListenAndServe(":8080", srv); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
