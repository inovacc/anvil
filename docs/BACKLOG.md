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

- **Priority:** P3
- **Description:** Secret templates for common patterns (database URLs, API keys)
- **Effort:** Small

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

### Technical Debt
- **Priority:** P1
- **Description:** Increase test coverage to 80%+ across all packages
- **Effort:** Medium
- **Status:** IN PROGRESS — Core packages at 72% (store 93%, vault 65%, crypto 74%, sentinel 72%). Bounded by untestable TPM branches.

- **Priority:** P2
- **Description:** Add integration tests for CLI commands
- **Effort:** Medium

- **Priority:** P3
- **Description:** Benchmark encryption/decryption performance
- **Effort:** Small
