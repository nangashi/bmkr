-- name: GetCustomer :one
SELECT id, name, email, password_hash, created_at, updated_at
FROM customers
WHERE id = $1;

-- name: CreateCustomer :one
INSERT INTO customers (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at, updated_at;
