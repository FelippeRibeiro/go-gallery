-- name: CriarConvite :one
INSERT INTO convites (token, id_album, email, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ObterConvitePorToken :one
SELECT * FROM convites WHERE token = $1;

-- name: MarcarConviteUsado :exec
UPDATE convites SET used_at = NOW() WHERE token = $1;

-- name: ListarConvitesPorAlbum :many
SELECT * FROM convites
WHERE id_album = $1
ORDER BY criado_em DESC;

-- name: revogarConvite :exec
DELETE FROM convites WHERE id = $1 AND id_album = $2;
