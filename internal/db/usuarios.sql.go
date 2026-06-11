package db

import (
	"context"
	"database/sql"
)

const usuarioCols = `id, nome, email, senha_hash, ativo, criado_em, tipo, foto_perfil, bio`

func scanUsuario(row interface{ Scan(...any) error }, i *Usuario) error {
	return row.Scan(
		&i.ID, &i.Nome, &i.Email, &i.SenhaHash, &i.Ativo, &i.CriadoEm, &i.Tipo,
		&i.FotoPerfil, &i.Bio,
	)
}

const criarFotografo = `
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'fotografo', TRUE)
RETURNING ` + usuarioCols

type CriarFotografoParams struct {
	Nome      string
	Email     string
	SenhaHash sql.NullString
}

func (q *Queries) CriarFotografo(ctx context.Context, arg CriarFotografoParams) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, criarFotografo, arg.Nome, arg.Email, arg.SenhaHash)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const obterFotografoPorEmail = `
SELECT ` + usuarioCols + ` FROM usuarios
WHERE email = $1 AND tipo = 'fotografo' AND ativo = TRUE LIMIT 1
`

func (q *Queries) ObterFotografoPorEmail(ctx context.Context, email string) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterFotografoPorEmail, email)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const obterFotografoPorID = `
SELECT ` + usuarioCols + ` FROM usuarios
WHERE id = $1 AND tipo = 'fotografo' AND ativo = TRUE
`

func (q *Queries) ObterFotografoPorID(ctx context.Context, id int64) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterFotografoPorID, id)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const criarCliente = `
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'cliente', TRUE)
RETURNING ` + usuarioCols

func (q *Queries) CriarCliente(ctx context.Context, arg CriarFotografoParams) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, criarCliente, arg.Nome, arg.Email, arg.SenhaHash)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const obterUsuarioPorEmail = `
SELECT ` + usuarioCols + ` FROM usuarios
WHERE email = $1 AND ativo = TRUE LIMIT 1
`

func (q *Queries) ObterUsuarioPorEmail(ctx context.Context, email string) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterUsuarioPorEmail, email)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const obterUsuarioPorID = `
SELECT ` + usuarioCols + ` FROM usuarios
WHERE id = $1 AND ativo = TRUE
`

func (q *Queries) ObterUsuarioPorID(ctx context.Context, id int64) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obterUsuarioPorID, id)
	var i Usuario
	return i, scanUsuario(row, &i)
}

const atualizarBioFotografo = `
UPDATE usuarios SET bio = $2 WHERE id = $1
`

func (q *Queries) AtualizarBioFotografo(ctx context.Context, id int64, bio string) error {
	_, err := q.db.ExecContext(ctx, atualizarBioFotografo, id, bio)
	return err
}

const atualizarFotoPerfilFotografo = `
UPDATE usuarios SET foto_perfil = $2 WHERE id = $1
`

func (q *Queries) AtualizarFotoPerfilFotografo(ctx context.Context, id int64, url string) error {
	_, err := q.db.ExecContext(ctx, atualizarFotoPerfilFotografo, id, url)
	return err
}

const listarFotografosComAlbumPublico = `
SELECT u.id, u.nome, u.foto_perfil, u.bio, COUNT(a.id) AS total_albuns
FROM usuarios u
INNER JOIN albuns a ON a.id_fotografo = u.id AND a.publico = TRUE AND a.ativo = TRUE
WHERE u.tipo = 'fotografo' AND u.ativo = TRUE
GROUP BY u.id, u.nome, u.foto_perfil, u.bio
ORDER BY total_albuns DESC
`

func (q *Queries) ListarFotografosComAlbumPublico(ctx context.Context) ([]FotografoPublico, error) {
	rows, err := q.db.QueryContext(ctx, listarFotografosComAlbumPublico)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FotografoPublico
	for rows.Next() {
		var i FotografoPublico
		if err := rows.Scan(&i.ID, &i.Nome, &i.FotoPerfil, &i.Bio, &i.TotalAlbuns); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
