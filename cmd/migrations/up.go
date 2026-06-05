package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/FelippeRibeiro/go-gallery/internal/db"
)

var schemaMigrationSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    id BIGSERIAL PRIMARY KEY,
		filename TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

func main() {
	_, conn, err := db.NewConn()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	defer conn.Close()
	_, err = conn.ExecContext(context.Background(), schemaMigrationSQL)
	if err != nil {
		log.Fatalf("Erro ao criar tabela de schema migrations: %v", err)
	}

	arquivos, err := os.ReadDir("db/schema")
	if err != nil {
		log.Fatalf("Erro ao ler diretório de migrations: %v", err)
	}

	var nomesArquivos []string

	for _, arquivo := range arquivos {
		if arquivo.IsDir() {
			continue
		}
		nomesArquivos = append(nomesArquivos, arquivo.Name())
	}

	sort.Strings(nomesArquivos)
	fmt.Println(nomesArquivos)

	for _, arquivo := range nomesArquivos {

		rows, err := conn.QueryContext(context.Background(), "SELECT 1 FROM schema_migrations WHERE filename = $1", arquivo)
		if err != nil {
			log.Fatalf("Erro ao verificar se a migration existe: %v", err)
		}
		if rows.Next() {
			fmt.Println("Migration já executada")
			continue
		}

		fmt.Println("Executando migration:", arquivo)
		sql, err := os.ReadFile(fmt.Sprintf("db/schema/%s", arquivo))
		if err != nil {
			log.Fatalf("Erro ao ler arquivo de migration: %v", err)
		}
		fmt.Println(string(sql))

		tx, err := conn.BeginTx(context.Background(), nil)
		if err != nil {
			log.Fatalf("Erro ao iniciar transação: %v", err)
		}
		_, err = tx.ExecContext(context.Background(), string(sql))
		if err != nil {
			log.Fatalf("Erro ao executar migration: %v", err)
			if err := tx.Rollback(); err != nil {
				log.Fatalf("Erro ao executar rollback: %v", err)
			}

		}
		if err := tx.Commit(); err != nil {
			log.Fatalf("Erro ao executar commit: %v", err)
		}
		fmt.Println("Migration executada com sucesso")
		_, err = conn.ExecContext(context.Background(), "INSERT INTO schema_migrations (filename) VALUES ($1)", arquivo)
		if err != nil {
			log.Fatalf("Erro ao inserir migration no banco de dados: %v", err)
		}
		fmt.Println("Migration inserida com sucesso")
	}
}
