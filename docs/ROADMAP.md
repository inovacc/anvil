# Project Roadmap

## Current Status
**Overall Progress:** 40% Complete

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

### Phase 4: Enhanced Security [NOT STARTED]
- [ ] Master key rotation
- [ ] Audit logging (access history)
- [ ] Secret versioning
- [ ] Backup and restore

### Phase 5: Integration [NOT STARTED]
- [ ] Shell integration helpers (bash, zsh, fish, powershell)
- [ ] Docker secrets bridge
- [ ] CI/CD pipeline integration
- [ ] Secret sharing between machines (encrypted export)

### Phase 6: Polish [NOT STARTED]
- [ ] Interactive TUI mode
- [ ] Auto-completion for shells
- [ ] Secret templates
- [ ] Plugin system
