-- name: CreatePost :one

INSERT INTO posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *
;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostByFeedID :many
SELECT * FROM posts WHERE feed_id = $1;

-- name: GetPostsByUser :many
SELECT * FROM posts WHERE feed_id IN (
    SELECT feed_id FROM feeds WHERE user_id = $1
) ORDER by published_at ASC LIMIT $2;
