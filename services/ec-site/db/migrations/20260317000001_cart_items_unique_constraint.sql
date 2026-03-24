-- +goose Up
ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_product_id_key UNIQUE (cart_id, product_id);
CREATE INDEX idx_carts_customer_id ON carts (customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_carts_customer_id;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_cart_id_product_id_key;
