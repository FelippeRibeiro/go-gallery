CREATE TABLE IF NOT EXISTS convites (
    id         BIGSERIAL PRIMARY KEY,
    token      TEXT        NOT NULL UNIQUE,
    id_album   BIGINT      NOT NULL REFERENCES albuns(id) ON DELETE CASCADE,
    email      TEXT        NOT NULL,
    expires_at TIMESTAMP   NOT NULL,
    used_at    TIMESTAMP   NULL,
    criado_em  TIMESTAMP   NOT NULL DEFAULT NOW()
);
