CREATE TABLE IF NOT EXISTS vault_audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,
    profile_name TEXT NOT NULL,
    secret_key TEXT DEFAULT '',
    detail TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_vault_audit_log_action ON vault_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_vault_audit_log_profile ON vault_audit_log(profile_name);
CREATE INDEX IF NOT EXISTS idx_vault_audit_log_created ON vault_audit_log(created_at);
