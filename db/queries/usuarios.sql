-- name: CriarFotografo :one
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'fotografo', TRUE)
RETURNING *;

-- name: ObterFotografoPorEmail :one
SELECT * FROM usuarios
WHERE email = $1 AND tipo = 'fotografo' AND ativo = TRUE LIMIT 1 ;

-- name: ObterFotografoPorID :one
SELECT * FROM usuarios
WHERE id = $1 AND tipo = 'fotografo' AND ativo = TRUE;

