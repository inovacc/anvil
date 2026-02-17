# anvil

Machine-bound encrypted vault and profile manager. Store and manage secrets organized by profiles, bound to the current
machine.

## Features

- **TPM 2.0 Hardware-Backed Sealing** — Master key sealed to TPM hardware when available, software HKDF fallback
  otherwise
- **Machine-Bound Encryption** — AES-256-GCM with HKDF-SHA256, non-portable by design
- **Profile Management** — Organize secrets into named profiles with default selection
- **Secret CRUD** — Set, get, delete, list, export, and import encrypted secrets
- **Secret Versioning** — Automatic version history with rollback capability
- **Master Key Rotation** — Rotate encryption key with transactional re-encryption of all secrets
- **Audit Logging** — Full action history with profile filtering
- **Backup & Restore** — Password-encrypted vault backups
- **Encrypted Sharing** — Cross-machine secret transfer with passphrase encryption
- **Secret Templates** — 5 built-in templates (postgres, mysql, redis, aws, github-token) with variable interpolation
- **Plugin System** — Event hooks (pre/post set, get, delete) and external secret providers
- **Secret Gathering** — Auto-discover secrets from `.env`, JSON, and YAML config files in a directory tree
- **Docker Bridge** — Export secrets as Docker files or Compose YAML snippets
- **Profile Isolation** — UUID-based `ScopedVault` for external apps: single-profile access with RBAC (masked reads,
  denied export)
- **Vault Seal/Unseal** — Temporarily lock all vault operations with `vault seal` / `vault unseal`
- **Version Lockdown** — Archived versions are metadata-only with 30-day retention; accessible only via rollback
- **Public Go API** — Clean `pkg/vault` module with interfaces (`VaultReader`, `VaultWriter`, `VaultEnv`,
  `VaultPassword`, `VaultScoped`) for external consumers
- **Password-Gated Env Release** — Time-limited secret access with bcrypt password gate
- **Multi-Format Export** — JSON, env, bash export, and PowerShell formats
- **Inline Secret Access** — Single secret retrieval via `--env-inline` flag
- **Global JSON Output** — Structured JSON output for all commands via `--json`
- **Interactive TUI** — Terminal UI for browsing profiles and managing secrets (`vault tui`)
- **User-Friendly Errors** — Clean error messages with actionable hints, no usage dump on errors
- **Memory Safety** — Master key zeroed on vault close via `sealbox.SecureZero`
- **Cross-Platform** — Windows (TPM via TBS), Linux (TPM via `/dev/tpmrm0`), macOS (software fallback)

## Installation

### Go Install

```bash
go install github.com/inovacc/anvil@latest
```

### From Source

```bash
git clone https://github.com/inovacc/anvil.git
cd anvil
task build
```

## Quick Start

```bash
# Initialize the vault
anvil vault init

# Create a profile
anvil vault profile create myapp --default

# Store secrets
anvil vault set API_KEY sk-abc123
anvil vault set DB_PASSWORD s3cret -d "Production database"

# Retrieve a secret
anvil vault get API_KEY

# List secrets
anvil vault list
```

## Interactive TUI

Launch an interactive terminal interface for vault management:

```bash
anvil vault tui
```

Navigate with keyboard shortcuts:

- **Dashboard**: vault status overview. `p` to browse profiles, `q` to quit
- **Profiles**: `↑`/`↓` to navigate, `Enter` to view secrets, `c` to create, `d` to delete, `u` to set default
- **Secrets**: `Enter` to reveal value, `n` to create, `d` to delete
- **Esc** goes back, **q** quits from any screen

## Vault Seal / Unseal

Temporarily lock all vault operations:

```bash
# Seal — blocks all read/write until unsealed
anvil vault seal

# Unseal — restores operations
anvil vault unseal
```

## Environment Variable Release

Release secrets as environment variables with password-gated, time-limited access:

```bash
# Set a password (min 8 characters)
anvil vault env password set

# Release secrets (default: 30 minutes)
anvil vault env release --ttl 1h

# Export as shell variables
eval $(anvil vault env export --format export)       # bash/zsh
anvil vault env export --format powershell | iex      # PowerShell

# Check release status
anvil vault env status

# Revoke access
anvil vault env revoke
```

## Import & Export

```bash
# Export secrets to JSON
anvil vault export --format json > secrets.json

# Export as env file
anvil vault export --format env > .env

# Import from file
anvil vault import secrets.json
anvil vault import .env --format env

# Gather secrets from a project directory
anvil vault gather ./myproject
anvil vault gather --yes --profile myapp ./myproject
```

## JSON Output

All commands support structured JSON output via the `--json` flag:

```bash
anvil vault status --json
anvil vault list --json
anvil vault get API_KEY --json
```

## Library Usage

Use `pkg/vault` as a Go library in your own applications:

```go
import "github.com/inovacc/anvil/pkg/vault"

// Open an existing vault
v, err := vault.Open(nil)
if err != nil {
    log.Fatal(err)
}
defer v.Close()

// Read a secret
value, err := v.Get("API_KEY", "myapp")
```

### Scoped Access for External Apps

Use `ScopedVault` when external apps need isolated, single-profile access with RBAC:

```go
// Open scoped access by profile UUID (Get returns masked values, Export is denied)
sv, err := vault.OpenScoped("profile-uuid-here", nil)
if err != nil {
    log.Fatal(err)
}
defer sv.Close()

sv.Set("API_KEY", "secret", "my api key")   // write allowed
val, _ := sv.Get("API_KEY")                  // returns "sec***key" (masked)
secrets, _ := sv.List()                       // metadata only
_, err = sv.Export()                          // returns ErrReadDenied
```

### Interfaces

Program against interfaces for testability:

```go
func NewService(reader vault.VaultReader) *Service {
    return &Service{vault: reader}
}
```

Available interfaces: `VaultReader` (read-only), `VaultWriter` (read+write), `VaultEnv` (env release), `VaultPassword` (
password ops), `VaultScoped` (isolated single-profile access), `VaultSeal` (seal/unseal).

## CLI Tools

```bash
# Display command tree
anvil cmdtree

# Generate AI-readable documentation
anvil aicontext
anvil aicontext --compact
anvil aicontext --category vault
```

## Command Reference

```
anvil
├── vault
│   ├── init              Initialize the encrypted vault
│   ├── status            Show vault status
│   ├── set <key> <val>   Set an encrypted secret
│   ├── get <key>         Get a decrypted secret value
│   ├── delete <key>      Delete a secret
│   ├── list              List secrets in a profile
│   ├── export            Export secrets in plaintext
│   ├── import <file>     Import secrets from a file
│   ├── gather [dir]      Discover and import secrets from files
│   ├── rotate-key        Rotate the vault master key
│   ├── audit             Show audit log entries
│   ├── history <key>     Show version history for a secret
│   ├── rollback <k> <v>  Rollback a secret to a previous version
│   ├── backup            Create encrypted vault backup
│   ├── restore           Restore vault from backup
│   ├── profile
│   │   ├── create        Create a new vault profile
│   │   ├── list          List all vault profiles
│   │   ├── delete        Delete a profile and its secrets
│   │   └── use           Set a profile as the default
│   ├── env
│   │   ├── password
│   │   │   ├── set       Set or update the vault env password
│   │   │   └── reset     Remove the vault env password
│   │   ├── release       Release secrets for a time-limited period
│   │   ├── revoke        Revoke the active env release
│   │   ├── status        Show current env release status
│   │   └── export        Export released secrets as env variables
│   ├── template
│   │   ├── create        Create a template from YAML/JSON
│   │   ├── list          List all templates
│   │   ├── show          Show template definition
│   │   ├── delete        Delete a template
│   │   └── apply         Apply template to create secrets
│   ├── plugin
│   │   ├── list          List configured hooks and providers
│   │   ├── hook-add      Add a hook for a vault event
│   │   ├── hook-remove   Remove a hook
│   │   ├── provider-add  Add a secret provider
│   │   └── provider-remove Remove a secret provider
│   ├── seal              Temporarily lock the vault
│   ├── unseal            Unlock a sealed vault
│   ├── share
│   │   ├── export        Export secrets with passphrase encryption
│   │   └── import        Import shared secrets
│   └── docker
│       ├── export        Write secrets as individual files
│       ├── clean         Remove exported secret files
│       └── compose       Generate Docker Compose YAML snippet
├── cmdtree               Display command tree
├── aicontext             Generate AI-readable documentation
└── completion            Shell completion scripts
```

## Security

- **TPM 2.0 sealing** — master key hardware-bound via [sealbox](https://github.com/inovacc/sealbox); cannot be extracted
  even with full disk access
- **Software fallback** — HKDF-SHA256 key derivation for machines without TPM (macOS, VMs)
- **AES-256-GCM** encryption for all stored secrets
- **bcrypt** password hashing for env release gate
- **Time-limited sessions** with automatic expiry for env release
- **Memory zeroing** — master key wiped from memory on vault close
- Secrets are **never cached on disk** in plaintext
- Database is **non-portable** — only works on the originating machine
- `vault status` shows current seal method (`tpm` or `software`)

## Development

```bash
task build          # Build to dist/
task test           # Run all tests with coverage
task lint           # Run golangci-lint
task check          # All quality checks (fmt, vet, lint, test)
task deps           # Download, tidy, verify dependencies
```

## License

See [LICENSE](LICENSE) for details.
