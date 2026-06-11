-- name: ListarClientesAlbum :many
SELECT ca.id, ca.id_cliente, ca.id_album, u.nome, u.email, ca.criado_em
FROM clientes_albuns ca
INNER JOIN usuarios u ON u.id = ca.id_cliente
WHERE ca.id_album = $1 AND u.ativo = TRUE
ORDER BY ca.criado_em ASC;

-- name: removerClienteAlbum :exec
DELETE FROM clientes_albuns WHERE id = $1 AND id_album = $2;
