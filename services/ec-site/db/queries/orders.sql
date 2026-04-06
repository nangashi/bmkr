-- name: CreateOrder :one
INSERT INTO orders (customer_id, total_amount, status)
VALUES ($1, $2, $3)
RETURNING id, customer_id, total_amount, status, created_at;

-- name: CreateOrderItem :exec
INSERT INTO order_items (order_id, product_id, product_name, price, quantity)
VALUES ($1, $2, $3, $4, $5);

-- name: ListOrdersByCustomerID :many
SELECT id, customer_id, total_amount, status, created_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: GetOrderByIDAndCustomerID :one
SELECT id, customer_id, total_amount, status, created_at
FROM orders
WHERE id = $1 AND customer_id = $2;

-- name: ListOrderItemsByOrderID :many
SELECT id, order_id, product_id, product_name, price, quantity
FROM order_items
WHERE order_id = $1;
