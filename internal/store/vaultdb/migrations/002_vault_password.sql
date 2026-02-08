-- Vault password table (singleton)
CREATE TABLE IF NOT EXISTS vault_password (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    password_hash BLOB NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
