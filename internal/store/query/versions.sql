-- name: InsertSecretVersion :exec
INSERT INTO vault_secret_versions (profile_name, key, version, encrypted_value, nonce)
VALUES (?, ?, ?, ?, ?);

-- name: ListSecretVersions :many
SELECT *
FROM vault_secret_versions
WHERE profile_name = ? AND key = ?
ORDER BY version DESC;

-- name: GetSecretVersion :one
SELECT *
FROM vault_secret_versions
WHERE profile_name = ? AND key = ? AND version = ? LIMIT 1;

-- name: GetLatestVersionNumber :one
SELECT CAST(COALESCE(MAX(version), 0) AS INTEGER) AS version
FROM vault_secret_versions
WHERE profile_name = ? AND key = ?;

-- name: DeleteSecretVersions :exec
DELETE
FROM vault_secret_versions
WHERE profile_name = ? AND key = ?;

-- name: UpdateSecretVersionData :exec
UPDATE vault_secret_versions
SET encrypted_value = ?,
    nonce           = ?
WHERE id = ?;

-- name: ListAllSecretVersions :many
SELECT *
FROM vault_secret_versions
ORDER BY profile_name, key, version;
