package db

import (
	"context"
	"database/sql"
)

const criarFotografo = `-- name: CriarFotografo :one
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'fotografo', TRUE)
RETURNING id, nome, email, senha_hash, ativo, criado_em, tipo
`

type CriarFotografoParams struct {
	Nome      string
	Email     string
	SenhaHash sql.NullString
}

func (q *Queries) CriarFotografo(ctx context.Context, arg CriarFotografoParams) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, criarFotografo, arg.Nome, arg.Email, arg.SenhaHash)
	var i Usuario
	err := row.Scan(
		&i.ID,
		&i.Nome,
		&i.Email,
		&i.SenhaHash,
		&i.Ativo,
		&i.CriadoEm,
		&i.Tipo,
	)
	return i, err
}

const obterFotografoPorEmail = `-- name: ObterFotografoPorEmail :one
SELECT id, nome, email, senha_hash, ativo, criado_em, tipo FROM usuarios
WHERE email = $1 AND tipo = 'fotografo' AND ativo = TRUE LIMIT 1
`

func (q *Queries) ObterFotografoPorEmail(ctx context.Context, email string) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterFotografoPorEmail, email)
	var i Usuario
	err := row.Scan(
		&i.ID,
		&i.Nome,
		&i.Email,
		&i.SenhaHash,
		&i.Ativo,
		&i.CriadoEm,
		&i.Tipo,
	)
	return i, err
}

const obterFotografoPorID = `-- name: ObterFotografoPorID :one
SELECT id, nome, email, senha_hash, ativo, criado_em, tipo FROM usuarios
WHERE id = $1 AND tipo = 'fotografo' AND ativo = TRUE
`

func (q *Queries) ObterFotografoPorID(ctx context.Context, id int64) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterFotografoPorID, id)
	var i Usuario
	err := row.Scan(
		&i.ID,
		&i.Nome,
		&i.Email,
		&i.SenhaHash,
		&i.Ativo,
		&i.CriadoEm,
		&i.Tipo,
	)
	return i, err
}

const criarCliente = `
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'cliente', TRUE)
RETURNING id, nome, email, senha_hash, ativo, criado_em, tipo
`

func (q *Queries) CriarCliente(ctx context.Context, arg CriarFotografoParams) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, criarCliente, arg.Nome, arg.Email, arg.SenhaHash)
	var i Usuario
	err := row.Scan(&i.ID, &i.Nome, &i.Email, &i.SenhaHash, &i.Ativo, &i.CriadoEm, &i.Tipo)
	return i, err
}

const obterUsuarioPorEmail = `
SELECT id, nome, email, senha_hash, ativo, criado_em, tipo FROM usuarios
WHERE email = $1 AND ativo = TRUE LIMIT 1
`

func (q *Queries) ObterUsuarioPorEmail(ctx context.Context, email string) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterUsuarioPorEmail, email)
	var i Usuario
	err := row.Scan(&i.ID, &i.Nome, &i.Email, &i.SenhaHash, &i.Ativo, &i.CriadoEm, &i.Tipo)
	return i, err
}

const obterUsuarioPorID = `
SELECT id, nome, email, senha_hash, ativo, criado_em, tipo FROM usuarios
WHERE id = $1 AND ativo = TRUE
`

func (q *Queries) ObterUsuarioPorID(ctx context.Context, id int64) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterUsuarioPorID, id)
	var i Usuario
	err := row.Scan(&i.ID, &i.Nome, &i.Email, &i.SenhaHash, &i.Ativo, &i.CriadoEm, &i.Tipo)
	return i, err
}
