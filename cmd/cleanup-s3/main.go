// Command cleanup-s3 reconcilia o bucket S3/MinIO com o banco de dados:
// lista todos os objetos do bucket e remove aqueles que NÃO estão referenciados
// em nenhuma coluna do banco que armazena uma key do S3
// (fotografias.url_alta/url_baixa, albuns.capa_url, usuarios.foto_perfil).
//
// Por segurança, roda em DRY-RUN por padrão (apenas lista os órfãos). Para
// apagar de fato, use a flag --apply.
//
//	go run ./cmd/cleanup-s3            # dry-run: só mostra o que seria apagado
//	go run ./cmd/cleanup-s3 --apply    # apaga os objetos órfãos
//
// Observação: o S3 não tem "pastas" de verdade — o que parece pasta é apenas um
// prefixo nas keys. Ao apagar todos os objetos órfãos, os prefixos vazios deixam
// de existir automaticamente.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
	s3client "github.com/FelippeRibeiro/go-gallery/internal/s3"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	apply := flag.Bool("apply", false, "apaga de fato os objetos órfãos (sem a flag, roda em dry-run)")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] .env não encontrado, usando variáveis do ambiente do sistema")
	}

	ctx := context.Background()

	if err := db.Init(); err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	referenciadas, err := coletarKeysReferenciadas(ctx)
	if err != nil {
		log.Fatalf("erro ao coletar keys do banco: %v", err)
	}
	log.Printf("Keys referenciadas no banco: %d", len(referenciadas))

	s3client.ObterS3Client()
	objetos, err := listarObjetosS3(ctx)
	if err != nil {
		log.Fatalf("erro ao listar objetos do S3: %v", err)
	}
	log.Printf("Objetos no bucket %q: %d", s3client.BucketName, len(objetos))

	var orfaos []string
	for _, key := range objetos {
		if !referenciadas[key] {
			orfaos = append(orfaos, key)
		}
	}

	if len(orfaos) == 0 {
		log.Println("Nenhum objeto órfão encontrado — bucket e banco estão sincronizados.")
		return
	}

	log.Printf("Encontrados %d objeto(s) órfão(s) (não mapeados no banco):", len(orfaos))
	for _, key := range orfaos {
		fmt.Println("  - " + key)
	}

	if !*apply {
		log.Println("")
		log.Println("DRY-RUN: nada foi apagado. Rode novamente com --apply para remover os órfãos.")
		return
	}

	var apagados, falhas int
	for _, key := range orfaos {
		if err := s3client.DeleteObject(ctx, key); err != nil {
			log.Printf("[ERRO] falha ao apagar %q: %v", key, err)
			falhas++
			continue
		}
		log.Printf("apagado: %s", key)
		apagados++
	}
	log.Printf("Concluído: %d apagado(s), %d falha(s).", apagados, falhas)
}

// coletarKeysReferenciadas devolve o conjunto de keys do S3 referenciadas no
// banco. Os valores são normalizados com KeyFromURL para tolerar registros
// legados que guardaram a URL completa em vez de apenas a key.
func coletarKeysReferenciadas(ctx context.Context) (map[string]bool, error) {
	conn := db.GetDB()
	keys := make(map[string]bool)

	adicionar := func(v sql.NullString) {
		if v.Valid && v.String != "" {
			keys[s3client.KeyFromURL(v.String)] = true
		}
	}

	// fotografias: original (url_alta) + preview (url_baixa)
	if err := scanColunas(ctx, conn,
		"SELECT url_alta, url_baixa FROM fotografias",
		func(rows *sql.Rows) error {
			var alta, baixa sql.NullString
			if err := rows.Scan(&alta, &baixa); err != nil {
				return err
			}
			adicionar(alta)
			adicionar(baixa)
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("fotografias: %w", err)
	}

	// albuns: capa
	if err := scanColunas(ctx, conn,
		"SELECT capa_url FROM albuns",
		func(rows *sql.Rows) error {
			var capa sql.NullString
			if err := rows.Scan(&capa); err != nil {
				return err
			}
			adicionar(capa)
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("albuns: %w", err)
	}

	// usuarios: foto de perfil
	if err := scanColunas(ctx, conn,
		"SELECT foto_perfil FROM usuarios",
		func(rows *sql.Rows) error {
			var foto sql.NullString
			if err := rows.Scan(&foto); err != nil {
				return err
			}
			adicionar(foto)
			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("usuarios: %w", err)
	}

	return keys, nil
}

// scanColunas executa a query e aplica scan a cada linha.
func scanColunas(ctx context.Context, conn *sql.DB, query string, scan func(*sql.Rows) error) error {
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

// listarObjetosS3 devolve todas as keys do bucket, paginando os resultados.
func listarObjetosS3(ctx context.Context) ([]string, error) {
	client := s3client.ObterS3Client()
	var keys []string
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3client.BucketName),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}
	return keys, nil
}
