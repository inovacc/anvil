package store

// AppSchema contains the CREATE TABLE statements for isolated app databases.
// App databases only store secrets — no profiles, sealed_key, audit, etc.
const AppSchema = `
CREATE TABLE IF NOT EXISTS app_secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    encrypted_value BLOB NOT NULL,
    nonce BLOB NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_app_secrets_key ON app_secrets(key);
`
