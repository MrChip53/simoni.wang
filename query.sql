-- name: GetLink :one
SELECT * FROM links WHERE id = $1;

-- name: CreateLink :one
INSERT INTO links (url)
VALUES ($1)
RETURNING *;