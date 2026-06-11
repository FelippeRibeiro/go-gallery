package s3client

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	instanciaS3 *s3.Client
	once        sync.Once
)

func ObterS3Client() *s3.Client {
	once.Do(func() {
		ctx := context.Background()

		minioEndpoint := os.Getenv("S3_ENDPOINT")
		if minioEndpoint == "" {
			minioEndpoint = "http://localhost:9000"
		}
		minioAccessKey := os.Getenv("S3_ACCESS_KEY")
		if minioAccessKey == "" {
			minioAccessKey = "admin"
		}
		minioSecretKey := os.Getenv("S3_SECRET_KEY")
		if minioSecretKey == "" {
			minioSecretKey = "xGGXvbprQad++HqXR3pGYhPddsF+pOr0wQep06TghnU="
		}

		credenciais := credentials.NewStaticCredentialsProvider(minioAccessKey, minioSecretKey, "")
		cfg, err := config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(credenciais),
			config.WithRegion("us-east-1"),
		)
		if err != nil {
			log.Fatalf("Erro crítico ao configurar SDK S3: %v", err)
		}

		instanciaS3 = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = &minioEndpoint
			o.UsePathStyle = true
		})

		fmt.Println("⚡ [S3] Conexão Singleton com MinIO inicializada com sucesso!")
	})
	return instanciaS3
}
