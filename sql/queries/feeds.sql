-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFeeds :many
SELECT name, url, user_id
FROM feeds
LIMIT 1000 OFFSET 0;


-- name: GetFeedByName :one
SELECT *
FROM feeds WHERE name = $1;
