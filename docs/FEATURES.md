# Features

## Completed

### Machine-Bound Encryption
- **Status:** Completed (v0.1.0)
- AES-256-GCM encryption for all stored secrets
- HKDF-SHA256 key derivation bound to machine hardware identity
- Sealed master key stored in database, non-portable by design

### Profile Management
- **Status:** Completed (v0.1.0)
- Create, list, delete named profiles
- Set a default profile for convenience
- Secrets organized per profile

### Secret CRUD
- **Status:** Completed (v0.1.0)
- Set, get, delete, list encrypted secrets
- Optional descriptions on secrets
- Secret existence checks

### Import & Export
- **Status:** Completed (v0.1.0)
- Export secrets as JSON, env, bash export, or PowerShell format
- Import from JSON or env files

### Password-Gated Environment Release
- **Status:** Completed (v0.2.0)
- bcrypt password hashing for vault access gate
- Time-limited secret release sessions (configurable TTL)
- File-based sentinel state management
- Auto-expiry and manual revoke
- Inline secret access via `--env-inline` flag

### Global JSON Output
- **Status:** Completed (v0.2.0)
- All commands support `--json` flag for structured output
- Dual output mode: human-readable text or machine-parseable JSON

### CLI Tooling
- **Status:** Completed (v0.2.0)
- Command tree visualization (`cmdtree`)
- AI-readable documentation generator (`aicontext`)
- Shell completion scripts

### TPM 2.0 Hardware-Backed Key Sealing
- **Status:** Completed (v0.3.0)
- Master key sealed to TPM 2.0 hardware via sealbox when available
- Transparent software fallback (HKDF) for machines without TPM
- `seal_method` column discriminates TPM vs software unseal path
- Platform support: Windows (TBS), Linux (`/dev/tpmrm0`); macOS falls back to software
- `vault status` displays current seal method

### Memory Safety
- **Status:** Completed (v0.3.0)
- Master key zeroed on vault close and after init via `sealbox.SecureZero`
- Master key only lives in memory during the vault lifecycle

### Sentinel Packed Encryption
- **Status:** Completed (v0.3.0)
- Sentinel files use `sealbox.Encrypt`/`sealbox.Decrypt` (packed nonce||ciphertext)
- Binary-compatible with the original manual nonce packing format

## Proposed

### Master Key Rotation
- **Status:** Proposed
- Allow rotating the master encryption key without re-creating the vault
- Re-encrypt all secrets with the new key

### Audit Logging
- **Status:** Proposed
- Track access history for secrets (who accessed what, when)
- Queryable audit log

### Secret Versioning
- **Status:** Proposed
- Keep history of secret values with rollback capability
- Version comparison

### Shell Integration Helpers
- **Status:** Proposed
- Native auto-completion for bash, zsh, fish, powershell
- Shell profile helpers for automatic env loading

### Docker Secrets Bridge
- **Status:** Proposed
- Bridge vault secrets into Docker container environments
- Integration with Docker Compose secrets

### Interactive TUI
- **Status:** Proposed
- Terminal UI for browsing and managing secrets
- Built with bubbletea

### Backup & Restore
- **Status:** Proposed
- Encrypted vault backup for disaster recovery
- Restore to same machine
