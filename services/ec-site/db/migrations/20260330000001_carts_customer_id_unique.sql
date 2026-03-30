-- +goose Up
DROP INDEX IF EXISTS idx_carts_customer_id;
ALTER TABLE carts ADD CONSTRAINT carts_customer_id_key UNIQUE (customer_id);

-- +goose Down
ALTER TABLE carts DROP CONSTRAINT IF EXISTS carts_customer_id_key;
CREATE INDEX idx_carts_customer_id ON carts (customer_id);
