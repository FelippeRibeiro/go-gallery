DO $$ BEGIN
    CREATE TYPE usuario_tipo AS ENUM ('fotografo', 'admin', 'cliente');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;


CREATE TABLE IF NOT EXISTS usuarios (
    id BIGSERIAL PRIMARY KEY,
    nome TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    senha_hash TEXT NULL,
    ativo BOOLEAN NOT NULL DEFAULT FALSE,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    tipo usuario_tipo NOT NULL DEFAULT 'cliente'
);