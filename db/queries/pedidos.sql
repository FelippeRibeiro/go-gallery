-- name: criarPedido :one
INSERT INTO pedidos (id_cliente, id_album, status, valor_total)
VALUES ($1, $2, 'pendente', $3)
RETURNING *;

-- name: criarPedidoFoto :exec
INSERT INTO pedidos_fotos (id_pedido, id_fotografia, valor_unitario)
VALUES ($1, $2, $3)
ON CONFLICT (id_pedido, id_fotografia) DO NOTHING;

-- name: ObterPedidoPorID :one
SELECT * FROM pedidos WHERE id = $1;

-- name: ListarPedidosPorCliente :many
SELECT * FROM pedidos
WHERE id_cliente = $1
ORDER BY criado_em DESC;

-- name: ListarFotosPedido :many
SELECT pf.id, pf.id_pedido, pf.id_fotografia, pf.valor_unitario,
       f.url_alta, f.url_baixa
FROM pedidos_fotos pf
INNER JOIN fotografias f ON f.id = pf.id_fotografia
WHERE pf.id_pedido = $1;

-- name: ListarPedidosResumoPorCliente :many
SELECT p.id, p.id_cliente, p.id_album, p.status, p.valor_total, p.criado_em,
       a.titulo AS album_titulo, a.capa_url,
       (SELECT COUNT(*) FROM pedidos_fotos pf WHERE pf.id_pedido = p.id) AS total_fotos
FROM pedidos p
INNER JOIN albuns a ON a.id = p.id_album
WHERE p.id_cliente = $1
ORDER BY p.criado_em DESC;
