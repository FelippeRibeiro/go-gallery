package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var conn *sql.DB
var queries *Queries

func NewConn() (*Queries, *sql.DB, error) {

	//TODO: pegar o dbURL do arquivo de configuração
	dbURL := "postgres://postgres:go-gallery@localhost:5432/gallery?sslmode=disable"
	var err error
	conn, err = sql.Open("pgx", dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao conectar ao banco de dados: %w", err)
	}
	err = conn.Ping()
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao pingar o banco de dados: %w", err)
	}
	db := New(conn)
	queries = db

	return db, conn, nil
}
