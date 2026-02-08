-- name: UpsertSecret :exec
INSERT INTO vault_secrets (profile_name, key, encrypted_value, nonce, description)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(profile_name, key) DO UPDATE SET
    encrypted_value = excluded.encrypted_value,
    nonce = excluded.nonce,
    description = excluded.description,
    updated_at = CURRENT_TIMESTAMP;

-- name: GetSecret :one
SELECT * FROM vault_secrets WHERE profile_name = ? AND key = ? LIMIT 1;

-- name: ListSecrets :many
SELECT * FROM vault_secrets WHERE profile_name = ? ORDER BY key;

-- name: DeleteSecret :exec
DELETE FROM vault_secrets WHERE profile_name = ? AND key = ?;

-- name: DeleteAllSecrets :exec
DELETE FROM vault_secrets WHERE profile_name = ?;

-- name: SecretExists :one
SELECT COUNT(*) FROM vault_secrets WHERE profile_name = ? AND key = ?;

-- name: CountSecrets :one
SELECT COUNT(*) FROM vault_secrets WHERE profile_name = ?;
