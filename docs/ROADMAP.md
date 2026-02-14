# Project Roadmap

## Current Status
**Overall Progress:** 75% Complete

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

### Phase 5: Enhanced Security [COMPLETE]
- [x] Master key rotation (`vault rotate-key` with transactional re-encryption)
- [x] Audit logging (action, profile, key, detail, timestamp)
- [x] Secret versioning with rollback
- [ ] Backup and restore

### Phase 6: Integration [IN PROGRESS]
- [x] Shell auto-completion (bash, zsh, fish, powershell) with dynamic profile/secret completion
- [x] Docker secrets bridge (`vault docker export/clean/compose`)
- [x] CI/CD pipeline integration (GitHub Actions: release on tag, test on PR, build on main)
- [ ] Secret sharing between machines (encrypted export)

### Phase 7: Polish [IN PROGRESS]
- [x] User-friendly error handling (UserError type, no usage dump, JSON error output)
- [ ] Interactive TUI mode
- [x] Auto-completion for shells
- [ ] Secret templates
- [ ] Plugin system
- [x] Release automation with goreleaser (GitHub Actions workflow)

## Test Coverage

**Current:** 72% (core packages)  |  **Target:** 80%

| Package | Coverage | Status |
|---------|----------|--------|
| internal/crypto | 73.7% | Good |
| internal/sentinel | 72.4% | Good |
| internal/store | 93.0% | Excellent |
| internal/store/sqlc | N/A | Generated code |
| internal/application | 0.0% | No tests |
| pkg/vault | 64.5% | Good (TPM branches untestable) |
| cmd | 0.0% | No tests (commands use real vault paths) |
