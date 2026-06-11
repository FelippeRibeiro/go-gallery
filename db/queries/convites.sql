-- name: CriarConvite :one
INSERT INTO convites (token, id_album, email, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, token, id_album, email, expires_at, used_at, criado_em;

-- name: ObterConvitePorToken :one
SELECT id, token, id_album, email, expires_at, used_at, criado_em
FROM convites
WHERE token = $1;

-- name: MarcarConviteUsado :exec
UPDATE convites SET used_at = NOW() WHERE token = $1;
