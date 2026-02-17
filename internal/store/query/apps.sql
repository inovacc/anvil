-- name: RegisterApp :exec
INSERT INTO vault_apps (uuid, name, description, service_id, db_path, status)
VALUES (?, ?, ?, ?, ?, 'active');

-- name: GetAppByUUID :one
SELECT * FROM vault_apps WHERE uuid = ? LIMIT 1;

-- name: GetAppByName :one
SELECT * FROM vault_apps WHERE name = ? LIMIT 1;

-- name: ListApps :many
SELECT * FROM vault_apps ORDER BY name;

-- name: ListActiveApps :many
SELECT * FROM vault_apps WHERE status = 'active' ORDER BY name;

-- name: UpdateAppStatus :exec
UPDATE vault_apps SET status = ? WHERE uuid = ?;

-- name: UpdateAppLastAccessed :exec
UPDATE vault_apps SET last_accessed_at = CURRENT_TIMESTAMP WHERE uuid = ?;

-- name: UpdateAppSecretCount :exec
UPDATE vault_apps SET secret_count = ? WHERE uuid = ?;

-- name: DeleteApp :exec
DELETE FROM vault_apps WHERE uuid = ?;
