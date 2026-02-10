package store

// Schema contains the CREATE TABLE statements for the vault database.
const Schema = `
-- Vault profiles table
CREATE TABLE IF NOT EXISTS vault_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT DEFAULT '',
    is_default INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_vault_profiles_name ON vault_profiles(name);
CREATE INDEX IF NOT EXISTS idx_vault_profiles_default ON vault_profiles(is_default);

-- Vault secrets table
CREATE TABLE IF NOT EXISTS vault_secrets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_name TEXT NOT NULL,
    key TEXT NOT NULL,
    encrypted_value BLOB NOT NULL,
    nonce BLOB NOT NULL,
    description TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_name, key),
    FOREIGN KEY (profile_name) REFERENCES vault_profiles(name) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vault_secrets_profile ON vault_secrets(profile_name);
CREATE INDEX IF NOT EXISTS idx_vault_secrets_key ON vault_secrets(profile_name, key);

-- Vault sealed key table (singleton)
CREATE TABLE IF NOT EXISTS vault_sealed_key (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    sealed_data BLOB NOT NULL,
    nonce BLOB NOT NULL,
    key_salt BLOB NOT NULL,
    version INTEGER DEFAULT 1,
    machine_id_hash BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Vault password table (singleton)
CREATE TABLE IF NOT EXISTS vault_password (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    password_hash BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
