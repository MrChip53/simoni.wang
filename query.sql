-- name: GetLink :one
SELECT * FROM links WHERE id = $1;

-- name: GetLinkBySlug :one
SELECT * FROM links WHERE slug = $1;

-- name: CreateLink :one
INSERT INTO links (url, slug)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteLink :exec
DELETE FROM links WHERE id = $1;

-- name: GetLastNLinks :many
SELECT * FROM links ORDER BY created_at DESC LIMIT $1;