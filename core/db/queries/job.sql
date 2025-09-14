-- name: GetJobByID :one
SELECT *
FROM jobs
WHERE id = $1;


-- name: CreateJob :one
INSERT INTO jobs DEFAULT
VALUES
RETURNING *;