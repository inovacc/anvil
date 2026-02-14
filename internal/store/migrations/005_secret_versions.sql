CREATE TABLE IF NOT EXISTS vault_secret_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_name TEXT NOT NULL,
    key TEXT NOT NULL,
    version INTEGER NOT NULL,
    encrypted_value BLOB NOT NULL,
    nonce BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (profile_name) REFERENCES vault_profiles(name) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_vault_secret_versions_key ON vault_secret_versions(profile_name, key);
CREATE UNIQUE INDEX IF NOT EXISTS idx_vault_secret_versions_unique ON vault_secret_versions(profile_name, key, version);
