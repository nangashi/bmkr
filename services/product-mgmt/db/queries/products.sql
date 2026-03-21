-- name: GetProduct :one
SELECT id, name, description, price, stock_quantity, created_at, updated_at
FROM products
WHERE id = $1;

-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock_quantity)
VALUES ($1, $2, $3, $4)
RETURNING id, name, description, price, stock_quantity, created_at, updated_at;

-- name: CountProducts :one
SELECT COUNT(*) FROM products;

-- name: ListProducts :many
SELECT id, name, description, price, stock_quantity, created_at, updated_at
FROM products
ORDER BY id;

-- name: ListProductsPaginated :many
SELECT id, name, description, price, stock_quantity, created_at, updated_at
FROM products
ORDER BY id
LIMIT $1 OFFSET $2;
