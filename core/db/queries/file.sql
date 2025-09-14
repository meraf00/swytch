-- name: GetFileByID :one
SELECT * FROM files WHERE id = $1;

-- name: CreateFile :one
INSERT INTO
    files (
        object_name,
        original_name,
        original_format
    )
VALUES ($1, $2, $3)
RETURNING
    *;