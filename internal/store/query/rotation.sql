-- name: ListAllSecrets :many
SELECT * FROM vault_secrets ORDER BY profile_name, key;
