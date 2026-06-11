-- name: verificarFotografoAlbum :one
SELECT EXISTS(
  SELECT 1 FROM fotografos_albuns WHERE id_fotografo = $1 AND id_album = $2
);

-- name: associarFotografoAlbum :one
INSERT INTO fotografos_albuns (id_fotografo, id_album)
VALUES ($1, $2)
ON CONFLICT (id_fotografo, id_album) DO NOTHING
RETURNING *;

-- name: ListarFotografosAlbum :many
SELECT fa.id, fa.id_fotografo, fa.id_album, u.nome, u.email, fa.criado_em
FROM fotografos_albuns fa
INNER JOIN usuarios u ON u.id = fa.id_fotografo
WHERE fa.id_album = $1 AND u.ativo = TRUE
ORDER BY fa.criado_em ASC;

-- name: removerFotografoAlbum :exec
DELETE FROM fotografos_albuns WHERE id = $1 AND id_album = $2;
