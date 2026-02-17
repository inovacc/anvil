-- name: CreateTemplate :exec
INSERT INTO vault_templates (name, description, template_data, builtin)
VALUES (?, ?, ?, ?);

-- name: GetTemplate :one
SELECT *
FROM vault_templates
WHERE name = ? LIMIT 1;

-- name: ListTemplates :many
SELECT *
FROM vault_templates
ORDER BY name;

-- name: DeleteTemplate :exec
DELETE
FROM vault_templates
WHERE name = ?;

-- name: TemplateExists :one
SELECT COUNT(*)
FROM vault_templates
WHERE name = ?;

-- name: UpdateTemplate :exec
UPDATE vault_templates
SET description   = ?,
    template_data = ?,
    updated_at    = CURRENT_TIMESTAMP
WHERE name = ?;
