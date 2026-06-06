CREATE TABLE IF NOT EXISTS albuns (
    id BIGSERIAL PRIMARY KEY,
    titulo TEXT NOT NULL,
    descricao TEXT NULL,

    data_evento DATE NOT NULL,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    id_fotografo BIGINT NOT NULL REFERENCES usuarios(id),

    lote BOOLEAN NOT NULL DEFAULT FALSE,
    valor_album DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    valor_unitario_fotografia DECIMAL(10, 2) NOT NULL DEFAULT 0.00
    
);

CREATE TABLE IF NOT EXISTS fotografos_albuns (
  id BIGSERIAL PRIMARY KEY,
  id_fotografo BIGINT NOT NULL REFERENCES usuarios(id),
  id_album BIGINT NOT NULL REFERENCES albuns(id),
  criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
  CONSTRAINT unique_fotografo_album UNIQUE (id_fotografo, id_album)
);

CREATE TABLE IF NOT EXISTS clientes_albuns (
    id BIGSERIAL PRIMARY KEY,

    id_cliente BIGINT NOT NULL REFERENCES usuarios(id),
    id_album BIGINT NOT NULL REFERENCES albuns(id),

    criado_em TIMESTAMP NOT NULL DEFAULT NOW(),

   CONSTRAINT unique_cliente_album UNIQUE (id_cliente, id_album)

);


CREATE TABLE IF NOT EXISTS fotografias (
    id BIGSERIAL PRIMARY KEY,

    url_alta TEXT NOT NULL,
    url_baixa TEXT NOT NULL,

    descricao TEXT NULL,

    id_fotografo BIGINT NOT NULL REFERENCES usuarios(id),
    id_album BIGINT NOT NULL REFERENCES albuns(id) ON DELETE CASCADE,

    valor_unitario DECIMAL(10, 2) NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    
    criado_em TIMESTAMP NOT NULL DEFAULT NOW()
);


-- CREATE TABLE IF NOT EXISTS compras_fotografias (
--     id BIGSERIAL PRIMARY KEY,
--     id_cliente BIGINT NOT NULL REFERENCES usuarios(id),
--     id_fotografia BIGINT NOT NULL REFERENCES fotografias(id),
--     pago BOOLEAN NOT NULL DEFAULT FALSE,
--     id_transacao_gateway TEXT NULL, 
--     pago_em TIMESTAMP NULL,
--     criado_em TIMESTAMP NOT NULL DEFAULT NOW(),
    
--     CONSTRAINT unique_compra_foto UNIQUE (id_cliente, id_fotografia)
-- );