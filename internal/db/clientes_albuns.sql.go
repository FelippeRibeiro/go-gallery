package db

import (
	"context"
	"time"
)

type ClienteAlbumInfo struct {
	ID        int64
	IDCliente int64
	IDAlbum   int64
	Nome      string
	Email     string
	CriadoEm  time.Time
}

const listarClientesAlbum = `
SELECT ca.id, ca.id_cliente, ca.id_album, u.nome, u.email, ca.criado_em
FROM clientes_albuns ca
INNER JOIN usuarios u ON u.id = ca.id_cliente
WHERE ca.id_album = $1 AND u.ativo = TRUE
ORDER BY ca.criado_em ASC
`

func (q *Queries) ListarClientesAlbum(ctx context.Context, idAlbum int64) ([]ClienteAlbumInfo, error) {
	rows, err := q.db.QueryContext(ctx, listarClientesAlbum, idAlbum)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ClienteAlbumInfo
	for rows.Next() {
		var i ClienteAlbumInfo
		if err := rows.Scan(&i.ID, &i.IDCliente, &i.IDAlbum, &i.Nome, &i.Email, &i.CriadoEm); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const removerClienteAlbum = `DELETE FROM clientes_albuns WHERE id = $1 AND id_album = $2`

func (q *Queries) RemoverClienteAlbum(ctx context.Context, id, idAlbum int64) error {
	_, err := q.db.ExecContext(ctx, removerClienteAlbum, id, idAlbum)
	return err
}
