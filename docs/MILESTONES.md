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
- **Test Coverage:** 13.6% (crypto 72.9%, sentinel 70.7%, vaultdb 0%, vault 0%, cmd 0%)

## v0.3.0 — Security Hardening (Planned)
- Master key rotation
- Audit logging
- Secret versioning
- **Coverage Target:** 80%+

## v0.4.0 — Shell Integration (Planned)
- Auto-completion for bash, zsh, fish, powershell
- Shell profile helpers
- Docker secrets bridge

## v1.0.0 — Production Ready (Planned)
- Full test coverage (80%+ all packages)
- Documentation complete
- CI/CD pipeline
- Release automation with goreleaser
