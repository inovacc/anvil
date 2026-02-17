# Milestones

## v0.1.0 — Foundation (Complete)
- Machine-bound encrypted vault
- Profile management
- Secret CRUD operations
- Import/Export functionality
- **Released:** v0.1.0, v0.1.1

## v0.2.0 — Password-Gated Env Release (Complete)
- Password-gated access to secrets
- Time-limited release sessions
- Multi-format env export (env, export, powershell)
- Inline secret access for shell substitution
- Global `--json` flag, cmdtree, and aicontext commands

## v0.3.0 — TPM Hardware Security (Complete)
- TPM 2.0 hardware-backed master key sealing via sealbox
- TPM-first init with transparent software fallback
- `seal_method` column in `vault_sealed_key` (`"tpm"` or `"software"`)
- Sentinel encryption migrated to sealbox packed format (binary-compatible)
- Memory zeroing on vault close and after init (`sealbox.SecureZero`)
- Idempotent migration for existing databases
- Vault status displays seal method
- GitHub Actions CI/CD: release workflow (goreleaser on tag push), test workflow (lint + vuln on PR), build workflow (Linux + Windows)
- **Released:** v0.3.0
- **Test Coverage:** crypto 75.8%, sentinel 73.7%, store 83.2%

## v0.3.x — Public API Boundary (Complete)
- Clean `pkg/vault` module with no internal type leaks in public signatures
- Vault-owned `ReleaseState` type (replaces leaked `internal/sentinel.ReleaseState`)
- Removed `MasterKey()` method (security exposure, zero callers)
- Interfaces: `VaultReader`, `VaultWriter`, `VaultEnv`, `VaultPassword` with compile-time checks
- Updated package documentation with interface usage examples

## v0.4.0 — Security Hardening (Complete)
- Master key rotation with transactional re-encryption
- Audit logging (all CRUD, rotation, env operations)
- Secret versioning with rollback
- Backup and restore (encrypted, password-protected)
- Encrypted secret sharing between machines

## v0.5.0 — Shell Integration (Complete)
- Auto-completion for bash, zsh, fish, powershell
- Docker secrets bridge (`vault docker export/clean/compose`)
- Secret templates (5 built-in: postgres, mysql, redis, aws, github-token)
- CLI integration tests via `ANVIL_DB_PATH` env var (15 tests)
- Plugin system (event hooks, secret providers, custom commands)
- Encryption/decryption benchmarks
- CI/CD coverage reporting with threshold enforcement
- **Released:** v0.5.0

## v0.6.0 — Access Control (Complete)
- UUID-based profile isolation (`ScopedVault`)
- Vault seal/unseal for temporary lockdown
- Version history lockdown (metadata-only, rollback-only, 30-day retention)

## v0.7.0 — Interactive TUI (Complete)
- Interactive TUI mode with bubbletea/lipgloss/bubbles
- Dashboard, profile browser, secrets table view (Key/Description/Created/Updated)
- Secret form for creating secrets and profiles
- **Released:** v0.7.0
- **Test Coverage:** store 95.7%, sentinel 78.6%, crypto 75.8%, vault 75.9%, application 72.7%, cmd 71.3%, tui 50.5%

## v1.0.0 — Production Ready (Planned)
- Full test coverage (80%+ all packages)
- Documentation complete
- CI/CD pipeline (release workflow done; expand test/build coverage)
- Release automation with goreleaser (done)
