-- name: InsertKey :exec
INSERT INTO vault_keys (name, algorithm, encrypted_private_key, nonce, public_key, fingerprint, description)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetKey :one
SELECT *
FROM vault_keys
WHERE name = ? AND algorithm = ? LIMIT 1;

-- name: GetKeyByName :one
SELECT *
FROM vault_keys
WHERE name = ? LIMIT 1;

-- name: ListKeys :many
SELECT *
FROM vault_keys
ORDER BY name, algorithm;

-- name: DeleteKey :exec
DELETE
FROM vault_keys
WHERE name = ? AND algorithm = ?;

-- name: DeleteKeyByName :exec
DELETE
FROM vault_keys
WHERE name = ?;

-- name: KeyExists :one
SELECT COUNT(*)
FROM vault_keys
WHERE name = ? AND algorithm = ?;

-- name: KeyExistsByName :one
SELECT COUNT(*)
FROM vault_keys
WHERE name = ?;

-- name: ListAllKeys :many
SELECT *
FROM vault_keys
ORDER BY name, algorithm;
