package db

import (
	"context"
	"time"
)

type Pedido struct {
	ID         int64
	IDCliente  int64
	IDAlbum    int64
	Status     string
	ValorTotal string
	CriadoEm  time.Time
}

type PedidoFotoInfo struct {
	ID            int64
	IDPedido      int64
	IDFotografia  int64
	ValorUnitario string
	UrlAlta       string
	UrlBaixa      string
}

const criarPedido = `
INSERT INTO pedidos (id_cliente, id_album, status, valor_total)
VALUES ($1, $2, 'pendente', $3)
RETURNING id, id_cliente, id_album, status, valor_total, criado_em
`

func (q *Queries) CriarPedido(ctx context.Context, idCliente, idAlbum int64, valorTotal string) (Pedido, error) {
	row := q.db.QueryRowContext(ctx, criarPedido, idCliente, idAlbum, valorTotal)
	var p Pedido
	err := row.Scan(&p.ID, &p.IDCliente, &p.IDAlbum, &p.Status, &p.ValorTotal, &p.CriadoEm)
	return p, err
}

const criarPedidoFoto = `
INSERT INTO pedidos_fotos (id_pedido, id_fotografia, valor_unitario)
VALUES ($1, $2, $3)
ON CONFLICT (id_pedido, id_fotografia) DO NOTHING
`

func (q *Queries) CriarPedidoFoto(ctx context.Context, idPedido, idFotografia int64, valorUnitario string) error {
	_, err := q.db.ExecContext(ctx, criarPedidoFoto, idPedido, idFotografia, valorUnitario)
	return err
}

const obterPedidoPorID = `
SELECT id, id_cliente, id_album, status, valor_total, criado_em
FROM pedidos WHERE id = $1
`

func (q *Queries) ObterPedidoPorID(ctx context.Context, id int64) (Pedido, error) {
	row := q.db.QueryRowContext(ctx, obterPedidoPorID, id)
	var p Pedido
	err := row.Scan(&p.ID, &p.IDCliente, &p.IDAlbum, &p.Status, &p.ValorTotal, &p.CriadoEm)
	return p, err
}

const listarFotosPedido = `
SELECT pf.id, pf.id_pedido, pf.id_fotografia, pf.valor_unitario, f.url_alta, f.url_baixa
FROM pedidos_fotos pf
INNER JOIN fotografias f ON f.id = pf.id_fotografia
WHERE pf.id_pedido = $1
`

func (q *Queries) ListarFotosPedido(ctx context.Context, idPedido int64) ([]PedidoFotoInfo, error) {
	rows, err := q.db.QueryContext(ctx, listarFotosPedido, idPedido)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PedidoFotoInfo
	for rows.Next() {
		var i PedidoFotoInfo
		if err := rows.Scan(&i.ID, &i.IDPedido, &i.IDFotografia, &i.ValorUnitario, &i.UrlAlta, &i.UrlBaixa); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

const listarPedidosPorCliente = `
SELECT id, id_cliente, id_album, status, valor_total, criado_em
FROM pedidos WHERE id_cliente = $1 ORDER BY criado_em DESC
`

func (q *Queries) ListarPedidosPorCliente(ctx context.Context, idCliente int64) ([]Pedido, error) {
	rows, err := q.db.QueryContext(ctx, listarPedidosPorCliente, idCliente)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Pedido
	for rows.Next() {
		var p Pedido
		if err := rows.Scan(&p.ID, &p.IDCliente, &p.IDAlbum, &p.Status, &p.ValorTotal, &p.CriadoEm); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}
