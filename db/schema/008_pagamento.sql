ALTER TABLE pedidos
    ADD COLUMN IF NOT EXISTS mp_preference_id TEXT,
    ADD COLUMN IF NOT EXISTS mp_payment_id TEXT;
