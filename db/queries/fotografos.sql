-- name: CriarFotografo :one
INSERT INTO fotografos (nome, email, senha_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ObterFotografoPorEmail :one
SELECT * FROM fotografos
WHERE email = $1 LIMIT 1;