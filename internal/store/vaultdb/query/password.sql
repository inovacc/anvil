-- name: GetPassword :one
SELECT * FROM vault_password WHERE id = 1;

-- name: HasPassword :one
SELECT COUNT(*) FROM vault_password WHERE id = 1;

-- name: UpsertPassword :exec
INSERT INTO vault_password (id, password_hash, created_at, updated_at)
VALUES (1, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    password_hash = excluded.password_hash,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeletePassword :exec
DELETE FROM vault_password WHERE id = 1;
