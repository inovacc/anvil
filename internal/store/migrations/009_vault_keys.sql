-- Vault asymmetric keys table
CREATE TABLE IF NOT EXISTS vault_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    algorithm TEXT NOT NULL,
    encrypted_private_key BLOB NOT NULL,
    nonce BLOB NOT NULL,
    public_key BLOB NOT NULL,
    fingerprint TEXT NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, algorithm)
);

CREATE INDEX IF NOT EXISTS idx_vault_keys_name ON vault_keys(name);
CREATE INDEX IF NOT EXISTS idx_vault_keys_fingerprint ON vault_keys(fingerprint);
