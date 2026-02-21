# Backlog

## Priority Levels

| Priority | Timeline      |
|----------|---------------|
| P1       | First month   |
| P2       | First quarter |
| P3       | Future        |

## Items

### Security

- **Priority:** P1
- **Description:** Add master key rotation support
- **Effort:** Large
- **Status:** DONE — `vault rotate-key` command with transactional re-encryption

- **Priority:** P2
- **Description:** Implement audit logging for secret access
- **Effort:** Medium
- **Status:** DONE — `vault audit` command, logs all CRUD + rotation + env operations

- **Priority:** P2
- **Description:** Add secret versioning with rollback capability
- **Effort:** Large
- **Status:** DONE — `vault history <key>` and `vault rollback <key> <version>` commands

### Usability

- **Priority:** P1
- **Description:** Shell auto-completion (bash, zsh, fish, powershell)
- **Effort:** Medium
- **Status:** DONE — Dynamic completion for profile names and secret keys

- **Priority:** P2
- **Description:** Interactive TUI with bubbletea
- **Effort:** Large
- **Status:** DONE — `vault tui` command with dashboard, profile browser, secrets table view (
  Key/Description/Created/Updated columns), and secret form screens

- **Priority:** P3
- **Description:** Secret templates for common patterns (database URLs, API keys)
- **Effort:** Small
- **Status:** DONE — `vault template` commands (create, list, show, delete, apply), 5 built-in templates (postgres,
  mysql, redis, aws, github-token), variable interpolation via text/template

### Integration

- **Priority:** P2
- **Description:** Docker secrets bridge for container environments
- **Effort:** Medium
- **Status:** DONE — `vault docker export/clean/compose` commands

- **Priority:** P2
- **Description:** CI/CD pipeline integration (GitHub Actions, GitLab CI)
- **Effort:** Medium
- **Status:** DONE — Release (goreleaser on tag), test (lint + test on PR), build (on main)

- **Priority:** P3
- **Description:** Encrypted secret export for machine-to-machine transfer
- **Effort:** Large
- **Status:** DONE — `vault share export/import` commands with passphrase-based encryption

### Access Control

- **Priority:** P1
- **Description:** UUID-based profile isolation for external apps (ScopedVault)
- **Effort:** Large
- **Status:** DONE — `OpenScoped(uuid)` with masked reads, denied export, write-allowed RBAC

- **Priority:** P1
- **Description:** Vault seal/unseal for temporary lockdown
- **Effort:** Medium
- **Status:** DONE — `vault seal` / `vault unseal` with file-based marker, all operations blocked when sealed

- **Priority:** P1
- **Description:** Version history lockdown — metadata only, rollback-only access
- **Effort:** Medium
- **Status:** DONE — archived versions show only version+timestamp, 30-day retention with auto-purge

### Recovery & Portability

- **Priority:** P1
- **Description:** BIP-39 mnemonic recovery for vault master key
- **Effort:** Medium
- **Status:** DONE — `recover` and `recovery-phrase` commands; 24-word mnemonic generated at init, SHA-256 hash stored in
  `vault_recovery` table

- **Priority:** P1
- **Description:** Per-app isolated vault databases
- **Effort:** Large
- **Status:** DONE — `vault app register/list/remove/disable/enable/open` commands; dedicated SQLite databases per app

### Cryptography

- **Priority:** P1
- **Description:** Asymmetric key management with Ed25519 and ECDSA P-256
- **Effort:** Large
- **Status:** DONE — `key generate/list/delete/export/import`, `sign`, `verify` commands; encrypted private key storage in `vault_keys` table; PKCS8/PKIX PEM format; `VaultKeyManagement` and `VaultSigner` interfaces

### MCP Integration

- **Priority:** P1
- **Description:** MCP server exposing vault operations as tools (secrets CRUD, key management, signing, audit, env export) via Go SDK
- **Effort:** Large
- **Status:** DONE — `anvil mcp serve` command; 17 tools (secret CRUD, profiles, keys, signing, audit, status, installation ID) and 1 resource (`anvil://status`); in-memory transport tests

### Sync

- **Priority:** P3
- **Description:** Gossip-based profile synchronization across machines (needs design refinement)
- **Effort:** Large

### Technical Debt

- **Priority:** P1
- **Description:** Increase test coverage to 80%+ across all packages
- **Effort:** Medium
- **Status:** IN PROGRESS — Total 60.7%. output 81.8%, mcpserver 80.7%, cmd 80.0%, vault 79.5%, sentinel 78.6%, crypto 76.2%, application 72.7%, store 71.9%, tui 42.9%. Gaps: sqlc generated code (0%), TUI needs mocking, TPM/platform branches.

- **Priority:** P2
- **Description:** Add integration tests for CLI commands
- **Effort:** Medium
- **Status:** DONE — `cmd/cmd_test.go` with 15 integration tests (init, profiles, secrets, audit, versioning,
  export/import, templates, plugins, errors, E2E plugin hook)

- **Priority:** P3
- **Description:** Benchmark encryption/decryption performance
- **Effort:** Small
- **Status:** DONE — Go benchmarks for Encrypt/Decrypt, DeriveKey, SealMasterKey, and full vault Set/Get roundtrips
