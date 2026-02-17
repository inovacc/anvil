-- name: CreateProfile :exec
INSERT INTO vault_profiles (name, description, is_default)
VALUES (?, ?, ?);

-- name: GetProfileByName :one
SELECT *
FROM vault_profiles
WHERE name = ? LIMIT 1;

-- name: GetDefaultProfile :one
SELECT *
FROM vault_profiles
WHERE is_default = 1 LIMIT 1;

-- name: ListProfiles :many
SELECT *
FROM vault_profiles
ORDER BY name;

-- name: ProfileExists :one
SELECT COUNT(*)
FROM vault_profiles
WHERE name = ?;

-- name: DeleteProfile :exec
DELETE
FROM vault_profiles
WHERE name = ?;

-- name: ClearAllDefaultProfiles :exec
UPDATE vault_profiles
SET is_default = 0;

-- name: SetProfileDefault :exec
UPDATE vault_profiles
SET is_default = 1,
    updated_at = CURRENT_TIMESTAMP
WHERE name = ?;

-- name: UpdateProfileDescription :exec
UPDATE vault_profiles
SET description = ?,
    updated_at  = CURRENT_TIMESTAMP
WHERE name = ?;
