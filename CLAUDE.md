# anvil

## Project Overview

Go CLI application built with Cobra.

- **Module**: `github.com/inovacc/anvil`
- **Go version**: 1.25
- **Architecture**: Hexagonal/Clean (cmd/anvil/, internal/, pkg/)

## Build & Run

```bash
task build          # Build to dist/
task run            # Run the application
go run ./cmd/anvil  # Run directly
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
task lint           # golangci-lint (with --fix)
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
├── cmd/anvil/      # CLI commands (Cobra) — binary name: anvil
│   ├── main.go     # Entry point, rootCmd, --json and --env-inline flags
│   ├── register.go # Subcommand registration
│   ├── output.go   # JSON/text output helper (outputResult, tableWriter)
│   ├── errors.go   # User-friendly error formatting (handleError)
│   ├── helpers.go  # Shared CLI helpers (readPassword)
│   ├── cmdtree.go  # Command tree visualization
│   ├── aicontext.go # AI context documentation generator
│   ├── id.go       # Machine-bound installation ID command
│   ├── key.go      # Parent key command
│   ├── key_generate.go # key generate <name> [--algorithm] [--description]
│   ├── key_list.go  # key list (tabwriter output)
│   ├── key_delete.go # key delete <name> [--algorithm]
│   ├── key_export.go # key export <name> [--private] [--output]
│   ├── key_import.go # key import <name> <pem-file>
│   ├── sign.go     # sign --key <name> (--file | --string) [-o output]
│   ├── verify.go   # verify --key <name> (--file | --string) (--signature | --signature-file)
│   ├── vault_app.go # Per-app isolated vault commands
│   ├── vault_app_cmds.go # App subcommands (register, list, info, set, get, delete, export, import)
│   ├── mcp.go      # MCP server command (anvil mcp — stdio transport)
│   ├── vault_recover.go # BIP-39 mnemonic recovery command
│   ├── vault_recovery_phrase.go # Show recovery phrase command
│   └── vault_gather.go # Recursive secret discovery from .env/JSON/YAML files
├── internal/       # Private application code
│   ├── application/ # Application directory resolution (cross-platform)
│   ├── crypto/     # AES-256-GCM encryption, HKDF key derivation, TPM sealing, machine ID, BIP-39 mnemonic, installation ID, Ed25519/ECDSA-P256 asymmetric crypto
│   ├── mcpserver/  # MCP server (tools + resources backed by pkg/vault)
│   ├── sentinel/   # Time-limited release session management (sealbox packed encrypt)
│   ├── tui/        # Interactive TUI (bubbletea/lipgloss/bubbles — table views for all screens)
│   └── store/      # SQLite database store (mutex-protected ops)
│       ├── sqlc/   # Generated query code (sqlc generate)
│       └── vaultdb.go # Database operations wrapper
├── pkg/vault/      # Public vault API (types, interfaces, UserError, TPM-first init/open)
│   ├── iface.go    # VaultReader, VaultWriter, VaultEnv, VaultPassword, VaultAudit, VaultVersioning, VaultKeyRotation, VaultRecovery, VaultIdentity, VaultKeyManagement, VaultSigner interfaces
│   ├── keys.go     # Asymmetric key management (GenerateKey, ListKeys, DeleteKey, ExportKeyPEM, ImportKeyPEM)
│   ├── sign.go     # Digital signing and verification (Sign, Verify)
│   ├── identity.go # Machine-bound installation ID (deterministic SHA-256)
│   ├── audit.go    # Audit logging (best-effort, never blocks operations)
│   ├── versions.go # Secret versioning and rollback
│   ├── rotate.go   # Master key rotation with transactional re-encryption
│   └── plugin.go   # Plugin system: event hooks and secret providers
├── tests/pentest/  # Security penetration tests (crypto, auth, injection, access control)
├── docs/           # Documentation
├── Taskfile.yml    # Task runner configuration
├── .golangci.yml   # Linter configuration
├── .goreleaser.yaml # Release configuration
├── .github/workflows/ # CI/CD (release on tag, test on PR, build on main)
└── main.go         # Entry point (in cmd/anvil/)
```

## Key Dependencies

| Package                              | Purpose                                                                         |
|--------------------------------------|---------------------------------------------------------------------------------|
| `github.com/tyler-smith/go-bip39`    | BIP-39 mnemonic generation and validation for vault recovery                    |
| `github.com/inovacc/sealbox`         | TPM 2.0 hardware-backed key sealing, AES-256-GCM packed encrypt, memory zeroing |
| `github.com/spf13/cobra`             | CLI framework                                                                   |
| `github.com/charmbracelet/bubbletea` | TUI framework                                                                   |
| `github.com/charmbracelet/bubbles`   | TUI components (table, textinput)                                               |
| `github.com/charmbracelet/lipgloss`  | TUI styling                                                                     |
| `github.com/google/uuid`             | UUID generation for profiles and app registration                               |
| `golang.org/x/crypto`                | HKDF-SHA256, bcrypt                                                             |
| `gopkg.in/yaml.v3`                   | YAML parsing for gather, templates, and config files                            |
| `modernc.org/sqlite`                 | Pure Go SQLite (no CGO)                                                         |
| `github.com/modelcontextprotocol/go-sdk` | MCP server SDK (tools, resources, stdio/in-memory transports)               |

## Security Architecture

### Master Key Sealing

Vault init tries TPM 2.0 first, falls back to software HKDF:

- **TPM path** (`seal_method = "tpm"`): `sealbox.NewKeyManager()` → `km.SealKey(masterKey)` → JSON-serialized
  `SealedData` stored in `vault_sealed_key.sealed_data`; `nonce` and `key_salt` are NULL
- **Software path** (`seal_method = "software"`): `HKDF-SHA256(machineID, salt)` derives wrapping key → AES-256-GCM
  encrypts master key; `nonce` and `key_salt` stored alongside
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
- **Auto-migrate on open**: `store.Open()` detects the current schema version and applies pending migrations
  transparently
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
- `migrations/006_templates.sql` — templates table
- `migrations/007_vault_apps.sql` — app registry table
- `migrations/008_recovery.sql` — BIP-39 recovery verification table
- `migrations/009_vault_keys.sql` — Asymmetric key storage table (name, algorithm, encrypted private key, public key, fingerprint)
- Idempotent `ALTER TABLE` runs in `store.Open()` for pre-migration databases

### sqlc Workflow

Queries live in `internal/store/query/*.sql`, schema in `internal/store/migrations/*.sql`:

```bash
cd internal/store && sqlc generate
```

Regenerate after changing any `.sql` file. Generated code is in `internal/store/sqlc/`.

## Conventions

- Use `task` (Taskfile) for all automation
- Use `glix install` instead of `go install` for CLI tools; use `go install ./cmd/anvil` to install locally
- Table-driven tests, 80% coverage minimum
- Mute unused returns: `_, _ = fmt.Fprintln(w, output)`
- Use `log/slog` for structured logging
- All commands use `outputResult(cmd, jsonData, textFn)` for JSON/text dual output; `--json` → JSON, default → tabwriter
  table
- CLI list commands use `tableWriter(w)` from `cmd/anvil/output.go` for consistent `text/tabwriter` table output (header row
  in CAPS, tab-separated columns)
- Global `--json` persistent flag on rootCmd inherited by all subcommands
- Errors use `vault.UserError` (Message + Hint) for user-friendly output; `handleError` in `cmd/anvil/errors.go` formats
  them (text or JSON based on `--json`)
- `SilenceErrors` and `SilenceUsage` are set on rootCmd; Cobra does not dump usage on errors
- All TUI list screens use `bubbles/table`: secrets (4 cols), profiles (5 cols), audit (5 cols), history (2 cols); each
  has a `newXModel()` constructor, `xTableColumns()` for sizing, and resize handling in `Update()`
- TUI navigation: dashboard → `p` profiles → `enter` secrets → `h` history; dashboard → `a` audit; `esc` goes back
- `tableStyles()` in theme.go provides shared table styling for all TUI screens
- `visibleSubcommands()` in cmdtree.go filters hidden commands and "help"
- TPM tests use `t.Skip("TPM not available")` when `!sealbox.IsAvailable()`
- `pkg/vault` is the public module boundary — never expose `internal/` types in its signatures
- `pkg/vault/iface.go` has compile-time `var _ Interface = (*Vault)(nil)` checks — update when adding public methods
- `InitWithRecovery()` replaces `Init()` in CLI — generates BIP-39 mnemonic from master key, stores hash in `vault_recovery`
- `RecoverFromMnemonic()` is a package-level function (like `Init`) — converts mnemonic back to master key, re-seals to current machine
- `ShowRecoveryPhrase()` requires an open vault — converts live master key to mnemonic via `bip39.NewMnemonic`
- Only the SHA-256 hash of the mnemonic is stored in the DB — the mnemonic itself is never persisted
- `InstallationID()` is deterministic: `SHA-256(machine_id_hash || sealed_data)` — no extra storage, computed from existing `vault_sealed_key` data
- `anvil id` is a top-level command — opens vault, prints 64-char hex ID, supports `--json`
- `toReleaseState()` in `env.go` converts internal `sentinel.ReleaseState` to public `vault.ReleaseState`
- Audit logging is best-effort — `logAudit()` never returns errors, only logs via `slog.Error`
- Secret versioning: `Set` archives previous value; `Delete` removes version history
- Key rotation uses `rotateKeyTx()` helper to isolate transaction scope from audit logging (avoids mutex deadlock)
- Docker bridge: `vault docker export` writes one file per secret; `vault docker compose` generates YAML snippet
- Plugin system: `PluginManager` loaded in `Open()` from `plugins.json` alongside vault DB; hooks fire on Set/Get/Delete
  via pre/post events; pre-hooks can block operations by returning `{"allow":false}`; post-hook errors are logged but
  never block
- `PluginManager` mutation methods (`AddHook`, `AddProvider`, `RemoveHook`, `RemoveProvider`) return `error` and are nil-safe
- Plugin config (`plugins.json`) is separate from the vault DB — no schema migration needed
- Gather command: `vault gather [dir]` recursively discovers `.env`/`.env.*`, `.json`, `.yaml`/`.yml` files; extracts
  secret-pattern keys (password, token, api_key, etc.); interactive by default, `--yes -p <profile>` for non-interactive
- Sentinel `defaultCacheDir()` respects `ANVIL_DB_PATH` env var — ensures test isolation for sentinel files
- Integration tests reset `--json` persistent flag in `execCmd()` to prevent state leakage between tests
- Asymmetric keys: Ed25519 (default) and ECDSA P-256; private keys AES-256-GCM encrypted in `vault_keys` table; public keys stored unencrypted for verification without decryption
- `Sign` outputs base64-encoded signature; `Verify` returns `VerifyResult.Valid` bool, CLI exits 1 on invalid
- Key rotation re-encrypts `vault_keys` private keys alongside secrets in `rotateKeyTx()`
- PEM export/import uses PKCS8 (private) and PKIX (public) standard formats, interoperable with openssl
- `key generate`, `key list`, `key delete`, `key export`, `key import` are subcommands of `key` parent command
- `sign` and `verify` are top-level commands (not under `vault`)
- MCP server: `anvil mcp` starts stdio JSON-RPC server; `internal/mcpserver/server.go` creates server backed by `pkg/vault`
- MCP tools: `secret_get/set/delete/list`, `profile_list/create/delete`, `key_generate/list/delete/export`, `sign`, `verify`, `audit_log`, `vault_status`, `installation_id`
- MCP resources: `anvil://status` (vault status JSON)
- MCP tool handlers use typed input structs with `jsonschema` tags; return typed output structs (SDK auto-marshals)
- MCP tests use `mcp.NewInMemoryTransports()` with real vault (temp DB + `ANVIL_SKIP_TPM=1`)
- Excluded from MCP: password ops, key rotation, backup/restore, recovery phrase, seal/unseal, env release (security-sensitive)
