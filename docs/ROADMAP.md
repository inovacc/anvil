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

### Phase 8: Recovery & App Isolation [COMPLETE]

- [x] BIP-39 mnemonic recovery phrase (24-word) generated during vault init
- [x] `recover` command: restore vault on new machine from mnemonic
- [x] `recovery-phrase` command: show recovery words (password-gated)
- [x] SHA-256 mnemonic hash stored in `vault_recovery` table (mnemonic never persisted)
- [x] Per-app isolated vault databases (`vault app register/list/remove/disable/enable`)
- [x] App-scoped vault access with dedicated SQLite databases

### Phase 9: Asymmetric Key Management [COMPLETE]

- [x] Ed25519 and ECDSA P-256 key pair generation
- [x] Encrypted private key storage in `vault_keys` table (migration 009)
- [x] Key CRUD: generate, list, delete, export (PEM), import (PEM)
- [x] Digital signing (`sign` command) with base64 output
- [x] Signature verification (`verify` command) with exit code 0/1
- [x] PKCS8/PKIX PEM format for interoperability with openssl
- [x] Key rotation re-encrypts asymmetric private keys
- [x] `VaultKeyManagement` and `VaultSigner` interfaces with compile-time checks
- [x] Penetration tests for key management security (`tests/pentest/asymmetric_test.go`)

### Phase 10: MCP Server [COMPLETE]

- [x] MCP server exposing vault operations as tools (get/set/delete/list secrets, profiles, env)
- [x] Key management tools (generate, list, delete, export)
- [x] Signing and verification tools
- [x] Audit log tool with profile filtering
- [x] Resource endpoint for vault status (`anvil://status`)
- [x] Stdio transport for CLI integration (`anvil mcp serve`)
- [x] In-memory transport integration tests (17 tools, 1 resource)

## Test Coverage

**Target:** 80%

| Package              | Coverage | Status                                                       |
|----------------------|----------|--------------------------------------------------------------|
| internal/output      | 81.8%    | Good — above target                                          |
| internal/mcpserver   | 80.7%    | Good — above target                                          |
| cmd/anvil            | 80.0%    | Good — at target                                             |
| pkg/vault            | 79.5%    | Good — near target, bounded by TPM + platform branches       |
| internal/sentinel    | 78.6%    | Good — bounded by error path mocking                         |
| internal/crypto      | 76.2%    | Good — bounded by TPM branches, mnemonic funcs covered       |
| internal/application | 72.7%    | Good — bounded by platform-specific paths                    |
| internal/store       | 71.9%    | Good — near target                                           |
| internal/tui         | 42.9%    | Low — vault-dependent paths need mocking                     |
| internal/store/sqlc  | 0.0%     | Generated code (excluded from target)                        |
| **Total**            | **60.7%** | Below 80% target — sqlc generated code pulls down average   |
