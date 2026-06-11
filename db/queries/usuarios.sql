-- name: CriarFotografo :one
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'fotografo', TRUE)
RETURNING *;

-- name: CriarCliente :one
INSERT INTO usuarios (nome, email, senha_hash, tipo, ativo)
VALUES ($1, $2, $3, 'cliente', TRUE)
RETURNING *;

-- name: ObterFotografoPorEmail :one
SELECT * FROM usuarios
WHERE email = $1 AND tipo = 'fotografo' AND ativo = TRUE
LIMIT 1;

-- name: ObterFotografoPorID :one
SELECT * FROM usuarios
WHERE id = $1 AND tipo = 'fotografo' AND ativo = TRUE;

-- name: ObterUsuarioPorEmail :one
SELECT * FROM usuarios
WHERE email = $1 AND ativo = TRUE
LIMIT 1;

-- name: ObterUsuarioPorID :one
SELECT * FROM usuarios
WHERE id = $1 AND ativo = TRUE;

-- name: atualizarBioFotografo :exec
UPDATE usuarios SET bio = $2 WHERE id = $1;

-- name: atualizarFotoPerfilFotografo :exec
UPDATE usuarios SET foto_perfil = $2 WHERE id = $1;

-- name: ListarFotografosComAlbumPublico :many
SELECT u.id, u.nome, u.foto_perfil, u.bio, COUNT(a.id) AS total_albuns
FROM usuarios u
INNER JOIN albuns a ON a.id_fotografo = u.id AND a.publico = TRUE AND a.ativo = TRUE
WHERE u.tipo = 'fotografo' AND u.ativo = TRUE
GROUP BY u.id, u.nome, u.foto_perfil, u.bio
ORDER BY total_albuns DESC;
