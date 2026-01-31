-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFeeds :many
SELECT name, url, user_id
FROM feeds
LIMIT 1000 OFFSET 0;

-- name: GetFeedByName :many
SELECT *
FROM feeds WHERE name = $1;

-- name: GetFeedByUrl :one
SELECT *
FROM feeds WHERE url = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds 
ORDER BY last_fetched_at ASC NULLS FIRST;

-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = $2, updated_at = $3 WHERE id = $1;
