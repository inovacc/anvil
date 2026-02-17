-- Vault apps registry table
CREATE TABLE IF NOT EXISTS vault_apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    service_id TEXT DEFAULT '',
    db_path TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    secret_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_vault_apps_uuid ON vault_apps(uuid);
CREATE INDEX IF NOT EXISTS idx_vault_apps_name ON vault_apps(name);
CREATE INDEX IF NOT EXISTS idx_vault_apps_status ON vault_apps(status);
