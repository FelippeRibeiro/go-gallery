-- name: CriarAlbum :one
INSERT INTO albuns (titulo, descricao, data_evento, id_fotografo, lote, valor_album, valor_unitario_fotografia)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, titulo, descricao, data_evento, criado_em, ativo, id_fotografo, lote, valor_album, valor_unitario_fotografia;

-- name: ListarAlbunsPorFotografo :many
SELECT id, titulo, descricao, data_evento, criado_em, ativo, id_fotografo, lote, valor_album, valor_unitario_fotografia
FROM albuns
WHERE id_fotografo = $1 AND ativo = TRUE
ORDER BY criado_em DESC;

-- name: ObterAlbumPorID :one
SELECT id, titulo, descricao, data_evento, criado_em, ativo, id_fotografo, lote, valor_album, valor_unitario_fotografia
FROM albuns
WHERE id = $1 AND ativo = TRUE;

-- name: CriarFotografia :one
INSERT INTO fotografias (url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario, ativo, criado_em;

-- name: ListarFotografiasPorAlbum :many
SELECT id, url_alta, url_baixa, descricao, id_fotografo, id_album, valor_unitario, ativo, criado_em
FROM fotografias
WHERE id_album = $1 AND ativo = TRUE
ORDER BY criado_em DESC;

-- name: AssociarClienteAlbum :one
INSERT INTO clientes_albuns (id_cliente, id_album)
VALUES ($1, $2)
RETURNING id, id_cliente, id_album, criado_em;
