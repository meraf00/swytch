-- name: CreateTask :one
INSERT INTO
    tasks (
        file_id,
        job_id,
        target_format
    )
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetTaskByID :one
SELECT * FROM tasks WHERE id = $1;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET
    status = $2,
    started_at = COALESCE($3, started_at),
    completed_at = COALESCE($4, completed_at),
    converted_file_name = COALESCE($5, converted_file_name),
    error_message = COALESCE($6, error_message)
WHERE
    id = $1
RETURNING
    *;