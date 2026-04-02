-- name: GetCartByCustomerID :one
SELECT id, customer_id, created_at, updated_at
FROM carts
WHERE customer_id = $1;

-- name: CreateCartIfNotExists :exec
INSERT INTO carts (customer_id) VALUES ($1) ON CONFLICT (customer_id) DO NOTHING;

-- name: ListCartItems :many
SELECT id, cart_id, product_id, quantity, created_at
FROM cart_items
WHERE cart_id = $1;

-- name: UpsertCartItem :execrows
INSERT INTO cart_items (cart_id, product_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (cart_id, product_id) DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity;

-- name: RemoveCartItem :execrows
DELETE FROM cart_items WHERE id = $1 AND cart_id = $2;

-- name: UpdateCartItemQuantity :execrows
UPDATE cart_items SET quantity = $1 WHERE id = $2 AND cart_id = $3;

-- name: GetCartItem :one
SELECT id, cart_id, product_id, quantity, created_at
FROM cart_items
WHERE id = $1 AND cart_id = $2;
