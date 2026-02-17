-- name: GetSealedKey :one
SELECT *
FROM vault_sealed_key
WHERE id = 1;

-- name: HasSealedKey :one
SELECT COUNT(*)
FROM vault_sealed_key
WHERE id = 1;

-- name: UpsertSealedKey :exec
INSERT INTO vault_sealed_key (id, sealed_data, nonce, key_salt, version, machine_id_hash, seal_method, created_at,
                              updated_at)
VALUES (1, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) ON CONFLICT(id) DO
UPDATE SET
    sealed_data = excluded.sealed_data,
    nonce = excluded.nonce,
    key_salt = excluded.key_salt,
    version = excluded.version,
    machine_id_hash = excluded.machine_id_hash,
    seal_method = excluded.seal_method,
    updated_at = CURRENT_TIMESTAMP;

-- name: DeleteSealedKey :exec
DELETE
FROM vault_sealed_key
WHERE id = 1;
