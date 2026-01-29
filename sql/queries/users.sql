-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetUserByName :one
SELECT * FROM users
WHERE name = $1;

-- name: GetUserByID :one
SELECT id, name FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT name, CASE WHEN name=$1 THEN true ELSE false END AS is_logged_in
FROM users LIMIT 1000 OFFSET 0;
