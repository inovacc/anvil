# Backlog

## Priority Levels
| Priority | Timeline |
|----------|----------|
| P1 | First month |
| P2 | First quarter |
| P3 | Future |

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
- **Status:** DONE — `vault tui` command with dashboard, profile browser, secrets table view (Key/Description/Created/Updated columns), and secret form screens

- **Priority:** P3
- **Description:** Secret templates for common patterns (database URLs, API keys)
- **Effort:** Small
- **Status:** DONE — `vault template` commands (create, list, show, delete, apply), 5 built-in templates (postgres, mysql, redis, aws, github-token), variable interpolation via text/template

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

### Sync
- **Priority:** P3
- **Description:** Gossip-based profile synchronization across machines (needs design refinement)
- **Effort:** Large

### Technical Debt
- **Priority:** P1
- **Description:** Increase test coverage to 80%+ across all packages
- **Effort:** Medium
- **Status:** MOSTLY DONE — store 95.7%, sentinel 78.6%, crypto 75.8%, vault 74.0%, application 72.7%, cmd 69.3%. Remaining gaps bounded by TPM branches, platform-specific code, and error paths requiring mocking.

- **Priority:** P2
- **Description:** Add integration tests for CLI commands
- **Effort:** Medium
- **Status:** DONE — `cmd/cmd_test.go` with 15 integration tests (init, profiles, secrets, audit, versioning, export/import, templates, plugins, errors, E2E plugin hook)

- **Priority:** P3
- **Description:** Benchmark encryption/decryption performance
- **Effort:** Small
- **Status:** DONE — Go benchmarks for Encrypt/Decrypt, DeriveKey, SealMasterKey, and full vault Set/Get roundtrips
