package db

import (
	"context"
	"time"
)

type FotografoAlbumInfo struct {
	ID          int64
	IDFotografo int64
	IDAlbum     int64
	Nome        string
	Email       string
	CriadoEm   time.Time
}

const verificarFotografoAlbum = `SELECT COUNT(*) FROM fotografos_albuns WHERE id_fotografo = $1 AND id_album = $2`

func (q *Queries) VerificarFotografoAlbum(ctx context.Context, idFotografo, idAlbum int64) (bool, error) {
	var n int
	err := q.db.QueryRowContext(ctx, verificarFotografoAlbum, idFotografo, idAlbum).Scan(&n)
	return n > 0, err
}

const associarFotografoAlbum = `
INSERT INTO fotografos_albuns (id_fotografo, id_album)
VALUES ($1, $2)
ON CONFLICT (id_fotografo, id_album) DO NOTHING
RETURNING id, id_fotografo, id_album, criado_em
`

func (q *Queries) AssociarFotografoAlbum(ctx context.Context, idFotografo, idAlbum int64) (FotografosAlbun, error) {
	row := q.db.QueryRowContext(ctx, associarFotografoAlbum, idFotografo, idAlbum)
	var i FotografosAlbun
	err := row.Scan(&i.ID, &i.IDFotografo, &i.IDAlbum, &i.CriadoEm)
	return i, err
}

const listarFotografosAlbum = `
SELECT fa.id, fa.id_fotografo, fa.id_album, u.nome, u.email, fa.criado_em
FROM fotografos_albuns fa
INNER JOIN usuarios u ON u.id = fa.id_fotografo
WHERE fa.id_album = $1 AND u.ativo = TRUE
ORDER BY fa.criado_em ASC
`

func (q *Queries) ListarFotografosAlbum(ctx context.Context, idAlbum int64) ([]FotografoAlbumInfo, error) {
	rows, err := q.db.QueryContext(ctx, listarFotografosAlbum, idAlbum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FotografoAlbumInfo
	for rows.Next() {
		var i FotografoAlbumInfo
		if err := rows.Scan(&i.ID, &i.IDFotografo, &i.IDAlbum, &i.Nome, &i.Email, &i.CriadoEm); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const removerFotografoAlbum = `DELETE FROM fotografos_albuns WHERE id = $1 AND id_album = $2`

func (q *Queries) RemoverFotografoAlbum(ctx context.Context, id, idAlbum int64) error {
	_, err := q.db.ExecContext(ctx, removerFotografoAlbum, id, idAlbum)
	return err
}

const listarAlbunsColaborador = `
SELECT a.id, a.titulo, a.descricao, a.data_evento, a.criado_em, a.ativo, a.id_fotografo, a.lote, a.valor_album, a.valor_unitario_fotografia, a.publico, a.capa_url
FROM albuns a
INNER JOIN fotografos_albuns fa ON fa.id_album = a.id
WHERE fa.id_fotografo = $1 AND a.ativo = TRUE
ORDER BY a.criado_em DESC
`

func (q *Queries) ListarAlbunsColaborador(ctx context.Context, idFotografo int64) ([]Albun, error) {
	rows, err := q.db.QueryContext(ctx, listarAlbunsColaborador, idFotografo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Albun
	for rows.Next() {
		var i Albun
		if err := scanAlbun(rows, &i); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}
