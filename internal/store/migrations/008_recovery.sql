-- Recovery table for BIP-39 mnemonic verification
CREATE TABLE IF NOT EXISTS vault_recovery (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    mnemonic_hash BLOB NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
