-- Referência externa única do pedido (UUID), usada como external_reference no
-- Mercado Pago. O ID numérico não serve: após um reset do banco os IDs
-- recomeçam e colidem com pagamentos antigos da mesma conta MP.
ALTER TABLE pedidos
    ADD COLUMN IF NOT EXISTS referencia TEXT UNIQUE;
