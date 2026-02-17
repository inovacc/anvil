-- name: GetRecovery :one
SELECT *
FROM vault_recovery
WHERE id = 1;

-- name: HasRecovery :one
SELECT COUNT(*)
FROM vault_recovery
WHERE id = 1;

-- name: UpsertRecovery :exec
INSERT INTO vault_recovery (id, mnemonic_hash, enabled, created_at)
VALUES (1, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    mnemonic_hash = excluded.mnemonic_hash,
    enabled = excluded.enabled;

-- name: DeleteRecovery :exec
DELETE
FROM vault_recovery
WHERE id = 1;
