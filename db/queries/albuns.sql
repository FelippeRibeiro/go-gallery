-- name: CriarAlbum :one
INSERT INTO albuns (titulo, descricao, data_evento, id_fotografo, lote, valor_album, valor_unitario_fotografia, publico)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListarAlbunsPorFotografo :many
SELECT * FROM albuns
WHERE id_fotografo = $1 AND ativo = TRUE
ORDER BY criado_em DESC;

-- name: ObterAlbumPorID :one
SELECT * FROM albuns
WHERE id = $1 AND ativo = TRUE;

-- name: ListarAlbunsPorCliente :many
SELECT a.*
FROM albuns a
INNER JOIN clientes_albuns ca ON ca.id_album = a.id
WHERE ca.id_cliente = $1 AND a.ativo = TRUE
ORDER BY a.criado_em DESC;

-- name: ListarAlbunsPublicos :many
SELECT * FROM albuns
WHERE publico = TRUE AND ativo = TRUE
ORDER BY criado_em DESC;

-- name: ObterAlbumPublico :one
SELECT * FROM albuns
WHERE id = $1 AND publico = TRUE AND ativo = TRUE;

-- name: atualizarPublicoAlbum :one
UPDATE albuns SET publico = $2 WHERE id = $1
RETURNING *;

-- name: atualizarCapaAlbum :one
UPDATE albuns SET capa_url = $2 WHERE id = $1
RETURNING *;

-- name: ListarAlbunsColaborador :many
SELECT a.*
FROM albuns a
INNER JOIN fotografos_albuns fa ON fa.id_album = a.id
WHERE fa.id_fotografo = $1 AND a.ativo = TRUE
ORDER BY a.criado_em DESC;

-- name: ListarAlbunsPublicosPorFotografo :many
SELECT * FROM albuns
WHERE publico = TRUE AND ativo = TRUE AND id_fotografo = $1
ORDER BY criado_em DESC;

-- ─── Fotografias ──────────────────────────────────────────────────────────────

-- name: CriarFotografia :one
INSERT INTO fotografias (url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ObterFotografiaPorID :one
SELECT * FROM fotografias WHERE id = $1;

-- name: ListarFotografiasPorAlbum :many
SELECT * FROM fotografias
WHERE id_album = $1 AND ativo = TRUE
ORDER BY ordem ASC, id ASC;

-- name: listarFotografiasPorAlbumPaginado :many
SELECT * FROM fotografias
WHERE id_album = $1 AND ativo = TRUE
ORDER BY ordem ASC, id ASC
LIMIT $2 OFFSET $3;

-- name: ContarFotografiasPorAlbum :one
SELECT COUNT(*) FROM fotografias
WHERE id_album = $1 AND ativo = TRUE;

-- name: atualizarOrdemFotografia :exec
UPDATE fotografias SET ordem = $3 WHERE id = $1 AND id_album = $2;

-- name: DesativarFotografia :exec
UPDATE fotografias SET ativo = FALSE WHERE id = $1;

-- name: DeletarFotografia :exec
DELETE FROM fotografias WHERE id = $1;

-- name: ContarPedidosDaFotografia :one
SELECT COUNT(*) FROM pedidos_fotos WHERE id_fotografia = $1;

-- ─── Clientes ─────────────────────────────────────────────────────────────────

-- name: associarClienteAlbum :one
INSERT INTO clientes_albuns (id_cliente, id_album)
VALUES ($1, $2)
ON CONFLICT (id_cliente, id_album) DO NOTHING
RETURNING *;

-- name: verificarAcessoClienteAlbum :one
SELECT EXISTS(
  SELECT 1 FROM clientes_albuns WHERE id_cliente = $1 AND id_album = $2
);

-- name: verificarConvitePendente :one
SELECT EXISTS(
  SELECT 1 FROM convites
  WHERE email = $1 AND id_album = $2 AND used_at IS NULL AND expires_at > NOW()
);
