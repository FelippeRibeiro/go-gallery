CREATE TABLE pedidos (
    id BIGSERIAL PRIMARY KEY,
    id_cliente BIGINT NOT NULL REFERENCES usuarios(id),
    id_album BIGINT NOT NULL REFERENCES albuns(id),
    status TEXT NOT NULL DEFAULT 'pendente',
    valor_total DECIMAL(10,2) NOT NULL DEFAULT 0,
    criado_em TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE pedidos_fotos (
    id BIGSERIAL PRIMARY KEY,
    id_pedido BIGINT NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    id_fotografia BIGINT NOT NULL REFERENCES fotografias(id),
    valor_unitario DECIMAL(10,2) NOT NULL DEFAULT 0,
    UNIQUE(id_pedido, id_fotografia)
);
