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
- **Test Coverage:** 15.0% (crypto 73.7%, sentinel 72.4%, store 0%, vault 0%, cmd 0%)

## v0.4.0 — Security Hardening (Planned)
- Master key rotation
- Audit logging
- Secret versioning
- **Coverage Target:** 80%+

## v0.5.0 — Shell Integration (Planned)
- Auto-completion for bash, zsh, fish, powershell
- Shell profile helpers
- Docker secrets bridge

## v1.0.0 — Production Ready (Planned)
- Full test coverage (80%+ all packages)
- Documentation complete
- CI/CD pipeline (release workflow done; expand test/build coverage)
- Release automation with goreleaser (done)
