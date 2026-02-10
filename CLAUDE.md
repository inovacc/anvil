# profile

## Project Overview

Go CLI application built with Cobra.

- **Module**: `github.com/inovacc/profile`
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
profile/
├── cmd/            # CLI commands (Cobra)
│   ├── output.go   # JSON/text output helper (outputResult)
│   ├── cmdtree.go  # Command tree visualization
│   └── aicontext.go # AI context documentation generator
├── internal/       # Private application code
│   ├── crypto/     # AES-256-GCM encryption, HKDF key derivation, TPM sealing, machine ID
│   ├── sentinel/   # Time-limited release session management (sealbox packed encrypt)
│   └── store/
│       └── vaultdb/ # SQLite database store (sqlc-generated queries, mutex-protected ops)
├── pkg/vault/      # Public vault API (types, errors, TPM-first init/open)
├── docs/           # Documentation
├── Taskfile.yml    # Task runner configuration
├── .golangci.yml   # Linter configuration
├── .goreleaser.yaml # Release configuration
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

### Database Migrations

- `migrations/001_initial.sql` — profiles, secrets, sealed_key tables
- `migrations/002_vault_password.sql` — password table
- `migrations/003_seal_method.sql` — adds `seal_method` column to `vault_sealed_key`
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
- `visibleSubcommands()` in cmdtree.go filters hidden commands and "help"
- TPM tests use `t.Skip("TPM not available")` when `!sealbox.IsAvailable()`
