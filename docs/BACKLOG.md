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

- **Priority:** P2
- **Description:** Implement audit logging for secret access
- **Effort:** Medium

- **Priority:** P2
- **Description:** Add secret versioning with rollback capability
- **Effort:** Large

### Usability
- **Priority:** P1
- **Description:** Shell auto-completion (bash, zsh, fish, powershell)
- **Effort:** Medium

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

- **Priority:** P2
- **Description:** CI/CD pipeline integration (GitHub Actions, GitLab CI)
- **Effort:** Medium

- **Priority:** P3
- **Description:** Encrypted secret export for machine-to-machine transfer
- **Effort:** Large

### Technical Debt
- **Priority:** P1
- **Description:** Increase test coverage to 80%+ across all packages (currently 15.0%)
- **Effort:** Medium
- **Details:** `internal/store` (0%), `pkg/vault` (0%), `cmd` (0%) need tests; `crypto` (73.7%) and `sentinel` (72.4%) need improvement

- **Priority:** P2
- **Description:** Add integration tests for CLI commands
- **Effort:** Medium

- **Priority:** P3
- **Description:** Benchmark encryption/decryption performance
- **Effort:** Small

- **Priority:** P2
- **Description:** Populate LICENSE file (currently empty)
- **Effort:** Small
