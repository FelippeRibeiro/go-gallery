package db

import (
	"context"
	"time"
)

const criarConvite = `
INSERT INTO convites (token, id_album, email, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, token, id_album, email, expires_at, used_at, criado_em
`

type CriarConviteParams struct {
	Token     string
	IDAlbum   int64
	Email     string
	ExpiresAt time.Time
}

func (q *Queries) CriarConvite(ctx context.Context, arg CriarConviteParams) (Convite, error) {
	row := q.db.QueryRowContext(ctx, criarConvite, arg.Token, arg.IDAlbum, arg.Email, arg.ExpiresAt)
	var i Convite
	err := row.Scan(&i.ID, &i.Token, &i.IDAlbum, &i.Email, &i.ExpiresAt, &i.UsedAt, &i.CriadoEm)
	return i, err
}

const obterConvitePorToken = `
SELECT id, token, id_album, email, expires_at, used_at, criado_em
FROM convites
WHERE token = $1
`

func (q *Queries) ObterConvitePorToken(ctx context.Context, token string) (Convite, error) {
	row := q.db.QueryRowContext(ctx, obterConvitePorToken, token)
	var i Convite
	err := row.Scan(&i.ID, &i.Token, &i.IDAlbum, &i.Email, &i.ExpiresAt, &i.UsedAt, &i.CriadoEm)
	return i, err
}

const marcarConviteUsado = `UPDATE convites SET used_at = NOW() WHERE token = $1`

func (q *Queries) MarcarConviteUsado(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, marcarConviteUsado, token)
	return err
}

const listarConvitesPorAlbum = `
SELECT id, token, id_album, email, expires_at, used_at, criado_em
FROM convites WHERE id_album = $1
ORDER BY criado_em DESC
`

func (q *Queries) ListarConvitesPorAlbum(ctx context.Context, idAlbum int64) ([]Convite, error) {
	rows, err := q.db.QueryContext(ctx, listarConvitesPorAlbum, idAlbum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Convite
	for rows.Next() {
		var i Convite
		if err := rows.Scan(&i.ID, &i.Token, &i.IDAlbum, &i.Email, &i.ExpiresAt, &i.UsedAt, &i.CriadoEm); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const revogarConvite = `DELETE FROM convites WHERE id = $1 AND id_album = $2`

func (q *Queries) RevogarConvite(ctx context.Context, id, idAlbum int64) error {
	_, err := q.db.ExecContext(ctx, revogarConvite, id, idAlbum)
	return err
}
