-- name: GetCartByCustomerID :one
SELECT id, customer_id, created_at, updated_at
FROM carts
WHERE customer_id = $1;

-- name: CreateCart :one
INSERT INTO carts (customer_id)
VALUES ($1)
RETURNING id, customer_id, created_at, updated_at;

-- name: ListCartItems :many
SELECT id, cart_id, product_id, quantity, created_at
FROM cart_items
WHERE cart_id = $1;
