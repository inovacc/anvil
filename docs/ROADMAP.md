# Project Roadmap

## Current Status

**Overall Progress:** 100% Complete

## Storage Compatibility Rule (MANDATORY)

**The vault storage structure (SQLite schema, file layout, sealed key format) MUST be preserved across all versions.**
This is a non-negotiable requirement:

1. **Never drop or rename** existing tables, columns, or indexes
2. **Never change column types** or constraints on existing columns
3. **All schema changes** must be additive (new tables, new nullable columns, new indexes)
4. **Every schema change** requires a new numbered migration file (`migrations/NNN_*.sql`) using
   `CREATE TABLE IF NOT EXISTS`, `ALTER TABLE ... ADD COLUMN` (idempotent)
5. **Version detection**: Migrations must detect the current schema version and apply only what's needed — older vaults
   opened with newer anvil versions must auto-migrate transparently
6. **Sealed key format**: Changes to `vault_sealed_key` structure or sealed data encoding require a migration that reads
   the old format and converts in-place — never break existing sealed vaults
7. **Sentinel file format**: Any changes to the sentinel binary format must include backward-compatible reading of the
   old format
8. **Backup/restore**: Exported backups from any previous version must remain importable by newer versions
9. **Test coverage**: Every migration must have a test that creates a database in the pre-migration state and verifies
   the migration succeeds without data loss

Violating this rule means users lose access to their encrypted secrets. **There are no exceptions.**

## Phases

### Phase 1: Foundation [COMPLETE]

- [x] Project scaffolding with Cobra CLI
- [x] SQLite database with sqlc code generation
- [x] AES-256-GCM encryption engine
- [x] Machine-bound master key sealing (HKDF-SHA256)
- [x] Cross-platform machine ID detection (Windows, Linux, macOS)

### Phase 2: Core Vault [COMPLETE]

- [x] Vault init / open / close lifecycle
- [x] Profile management (create, list, delete, use)
- [x] Secret CRUD operations (set, get, delete, list)
- [x] Secret export (JSON, env format)
- [x] Secret import (JSON, env format)
- [x] Vault status reporting

### Phase 3: Password-Gated Environment Release [COMPLETE]

- [x] Password storage with bcrypt hashing
- [x] Sentinel-based session management (file-based state)
- [x] Time-limited secret release (configurable TTL)
- [x] Env export in multiple formats (env, export, powershell)
- [x] Inline secret access (`--env-inline`)
- [x] Auto-expiry and manual revoke

### Phase 4: TPM Hardware Security [COMPLETE]

- [x] TPM 2.0 hardware-backed master key sealing via sealbox
- [x] TPM-first init with transparent software fallback
- [x] `seal_method` discriminator in database (`tpm` or `software`)
- [x] Sentinel encryption migrated to sealbox packed format
- [x] Memory zeroing on vault close and after init
- [x] Idempotent database migration for `seal_method` column
- [x] Vault status displays seal method

### Phase 4b: Public API Boundary [COMPLETE]

- [x] Clean `pkg/vault` module with no internal type leaks
- [x] `ReleaseState` type owned by `pkg/vault` (was leaking `internal/sentinel`)
- [x] Removed `MasterKey()` security exposure from public API
- [x] Interfaces: `VaultReader`, `VaultWriter`, `VaultEnv`, `VaultPassword`
- [x] Compile-time interface satisfaction checks

### Phase 4c: Profile Isolation & Access Control [COMPLETE]

- [x] UUID-based profile identification (auto-generated on create)
- [x] ScopedVault for external app isolation (UUID-only access, single profile)
- [x] RBAC enforcement: scoped access returns masked values, denies plaintext export
- [x] `VaultScoped` interface with compile-time satisfaction check
- [x] Vault seal/unseal for temporary lockdown (`vault seal` / `vault unseal`)
- [x] Version history lockdown: metadata-only (no plaintext), rollback-only access
- [x] 30-day retention with automatic purge of expired versions

### Phase 5: Enhanced Security [COMPLETE]

- [x] Master key rotation (`vault rotate-key` with transactional re-encryption)
- [x] Audit logging (action, profile, key, detail, timestamp)
- [x] Secret versioning with rollback
- [x] Backup and restore

### Phase 6: Integration [COMPLETE]

- [x] Shell auto-completion (bash, zsh, fish, powershell) with dynamic profile/secret completion
- [x] Docker secrets bridge (`vault docker export/clean/compose`)
- [x] CI/CD pipeline integration (GitHub Actions: release on tag, test on PR, build on main)
- [x] Secret sharing between machines (encrypted export)
- [x] Secret gathering (`vault gather` — recursive discovery of .env and config files with secret-pattern extraction)

### Phase 7: Polish [COMPLETE]

- [x] User-friendly error handling (UserError type, no usage dump, JSON error output)
- [x] Interactive TUI mode (bubbletea/lipgloss — dashboard, profiles, secrets table view, secret form)
- [x] Auto-completion for shells
- [x] Secret templates (CRUD + apply with variable interpolation, 5 built-in templates)
- [x] Plugin system (event hooks, secret providers, custom commands via Go plugin interface)
- [x] Release automation with goreleaser (GitHub Actions workflow)
- [x] CLI integration tests (`cmd/cmd_test.go` via `ANVIL_DB_PATH` env var)
- [x] Encryption/decryption benchmarks

## Test Coverage

**Target:** 80%

| Package              | Coverage | Status                                                       |
|----------------------|----------|--------------------------------------------------------------|
| internal/store       | 95.7%    | Excellent                                                    |
| internal/sentinel    | 78.6%    | Good — bounded by error path mocking                         |
| internal/crypto      | 75.8%    | Good — bounded by TPM branches                               |
| pkg/vault            | 75.9%    | Good — bounded by TPM + platform branches                    |
| internal/application | 72.7%    | Good — bounded by platform-specific paths                    |
| cmd                  | 71.3%    | 40+ integration tests; bounded by slow crypto ops            |
| internal/tui         | 50.5%    | Fair — pure logic tested, vault-dependent paths need mocking |
| internal/store/sqlc  | N/A      | Generated code                                               |
