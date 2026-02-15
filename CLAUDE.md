# anvil

## Project Overview

Go CLI application built with Cobra.

- **Module**: `github.com/inovacc/anvil`
- **Go version**: 1.25
- **Architecture**: Hexagonal/Clean (cmd/, internal/, pkg/)

## Build & Run

```bash
task build          # Build to dist/
task run            # Run the application
go run .            # Run directly
```

## Testing

```bash
task test           # Run all tests with coverage
task test:unit      # Unit tests only (skip integration)
task test:cover     # Show coverage percentage
task test:coverage  # Generate HTML coverage report
```

## Code Quality

```bash
task fmt            # Format code (go fmt + goimports)
task vet            # Static analysis
task lint           # golangci-lint
task lint:fix       # Lint with auto-fix
task check          # All quality checks (fmt, vet, lint, test)
```

## Dependencies

```bash
task deps           # Download, tidy, verify
task deps:upgrade   # Upgrade all to latest
```

## Release

```bash
task build:dev          # Snapshot build with goreleaser
task release            # Production release (requires git tag)
task release:snapshot   # Snapshot release
task release:check      # Validate goreleaser config
```

## Project Structure

```
anvil/
├── cmd/            # CLI commands (Cobra)
│   ├── output.go   # JSON/text output helper (outputResult)
│   ├── errors.go   # User-friendly error formatting (handleError)
│   ├── cmdtree.go  # Command tree visualization
│   └── aicontext.go # AI context documentation generator
├── internal/       # Private application code
│   ├── application/ # Application directory resolution (cross-platform)
│   ├── crypto/     # AES-256-GCM encryption, HKDF key derivation, TPM sealing, machine ID
│   ├── sentinel/   # Time-limited release session management (sealbox packed encrypt)
│   └── store/      # SQLite database store (mutex-protected ops)
│       ├── sqlc/   # Generated query code (sqlc generate)
│       └── vaultdb.go # Database operations wrapper
├── pkg/vault/      # Public vault API (types, interfaces, UserError, TPM-first init/open)
│   ├── iface.go    # VaultReader, VaultWriter, VaultEnv, VaultPassword, VaultAudit, VaultVersioning, VaultKeyRotation interfaces
│   ├── audit.go    # Audit logging (best-effort, never blocks operations)
│   ├── versions.go # Secret versioning and rollback
│   ├── rotate.go   # Master key rotation with transactional re-encryption
│   └── plugin.go   # Plugin system: event hooks and secret providers
├── docs/           # Documentation
├── Taskfile.yml    # Task runner configuration
├── .golangci.yml   # Linter configuration
├── .goreleaser.yaml # Release configuration
├── .github/workflows/ # CI/CD (release on tag, test on PR, build on main)
└── main.go         # Entry point
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/inovacc/sealbox` | TPM 2.0 hardware-backed key sealing, AES-256-GCM packed encrypt, memory zeroing |
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/crypto` | HKDF-SHA256, bcrypt |
| `modernc.org/sqlite` | Pure Go SQLite (no CGO) |

## Security Architecture

### Master Key Sealing

Vault init tries TPM 2.0 first, falls back to software HKDF:

- **TPM path** (`seal_method = "tpm"`): `sealbox.NewKeyManager()` → `km.SealKey(masterKey)` → JSON-serialized `SealedData` stored in `vault_sealed_key.sealed_data`; `nonce` and `key_salt` are NULL
- **Software path** (`seal_method = "software"`): `HKDF-SHA256(machineID, salt)` derives wrapping key → AES-256-GCM encrypts master key; `nonce` and `key_salt` stored alongside
- The `seal_method` column discriminates which unseal path `Open()` uses
- TPM platform support: Linux (`/dev/tpmrm0`), Windows (TBS); macOS falls back to software

### Memory Safety

- `crypto.ZeroBytes()` (delegates to `sealbox.SecureZero()`) zeros master key on `Vault.Close()` and after `Init()`
- Master key only lives in memory during the vault lifecycle

### Sentinel Encryption

- Sentinel files use `sealbox.Encrypt`/`sealbox.Decrypt` (packed `nonce(12B) || ciphertext` format)
- Binary-compatible with the original manual nonce packing

### Storage Compatibility (MANDATORY RULE)

**Never break existing vault storage.** All schema/format changes must preserve backward compatibility:

- **No dropping/renaming** tables, columns, or indexes — only additive changes
- **Every schema change** gets a new numbered migration (`migrations/NNN_*.sql`) that is idempotent
- **Auto-migrate on open**: `store.Open()` detects the current schema version and applies pending migrations transparently
- **Sealed key & sentinel format changes** must read old formats and convert in-place
- **Backups from any prior version** must remain importable
- **Every migration must have a test** that creates a pre-migration DB and verifies data survives
- Violating this rule means users lose access to their encrypted secrets — **no exceptions**

### Database Migrations

- `migrations/001_initial.sql` — profiles, secrets, sealed_key tables
- `migrations/002_vault_password.sql` — password table
- `migrations/003_seal_method.sql` — adds `seal_method` column to `vault_sealed_key`
- `migrations/004_audit_log.sql` — audit log table
- `migrations/005_secret_versions.sql` — secret version history table
- Idempotent `ALTER TABLE` runs in `store.Open()` for pre-migration databases

### sqlc Workflow

Queries live in `internal/store/query/*.sql`, schema in `internal/store/migrations/*.sql`:

```bash
cd internal/store && sqlc generate
```

Regenerate after changing any `.sql` file. Generated code is in `internal/store/sqlc/`.

## Conventions

- Use `task` (Taskfile) for all automation
- Use `glix install` instead of `go install` for CLI tools; use `glix install .` to install from a local directory
- Table-driven tests, 80% coverage minimum
- Mute unused returns: `_, _ = fmt.Fprintln(w, output)`
- Use `log/slog` for structured logging
- All commands use `outputResult(cmd, jsonData, textFn)` for JSON/text dual output
- Global `--json` persistent flag on rootCmd inherited by all subcommands
- Errors use `vault.UserError` (Message + Hint) for user-friendly output; `handleError` in `cmd/errors.go` formats them (text or JSON based on `--json`)
- `SilenceErrors` and `SilenceUsage` are set on rootCmd; Cobra does not dump usage on errors
- `visibleSubcommands()` in cmdtree.go filters hidden commands and "help"
- TPM tests use `t.Skip("TPM not available")` when `!sealbox.IsAvailable()`
- `pkg/vault` is the public module boundary — never expose `internal/` types in its signatures
- `pkg/vault/iface.go` has compile-time `var _ Interface = (*Vault)(nil)` checks — update when adding public methods
- `toReleaseState()` in `env.go` converts internal `sentinel.ReleaseState` to public `vault.ReleaseState`
- Audit logging is best-effort — `logAudit()` never returns errors, only logs via `slog.Error`
- Secret versioning: `Set` archives previous value; `Delete` removes version history
- Key rotation uses `rotateKeyTx()` helper to isolate transaction scope from audit logging (avoids mutex deadlock)
- Docker bridge: `vault docker export` writes one file per secret; `vault docker compose` generates YAML snippet
- Plugin system: `PluginManager` loaded in `Open()` from `plugins.json` alongside vault DB; hooks fire on Set/Get/Delete via pre/post events; pre-hooks can block operations by returning `{"allow":false}`; post-hook errors are logged but never block
- Plugin config (`plugins.json`) is separate from the vault DB — no schema migration needed
