-- name: InsertAuditLog :exec
INSERT INTO vault_audit_log (action, profile_name, secret_key, detail)
VALUES (?, ?, ?, ?);

-- name: ListAuditLog :many
SELECT * FROM vault_audit_log ORDER BY created_at DESC LIMIT ?;

-- name: ListAuditLogByProfile :many
SELECT * FROM vault_audit_log WHERE profile_name = ? ORDER BY created_at DESC LIMIT ?;

-- name: ListAuditLogByAction :many
SELECT * FROM vault_audit_log WHERE action = ? ORDER BY created_at DESC LIMIT ?;

-- name: CountAuditLog :one
SELECT COUNT(*) FROM vault_audit_log;

-- name: PurgeAuditLog :exec
DELETE FROM vault_audit_log WHERE created_at < ?;
