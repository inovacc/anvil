# profile

Machine-bound encrypted vault and profile manager. Store and manage secrets organized by profiles, bound to the current machine.

## Features

- **Machine-Bound Encryption** — AES-256-GCM with HKDF-SHA256 master key sealed to hardware identity
- **Profile Management** — Organize secrets into named profiles with default selection
- **Secret CRUD** — Set, get, delete, list, export, and import encrypted secrets
- **Password-Gated Env Release** — Time-limited secret access with bcrypt password gate
- **Multi-Format Export** — JSON, env, bash export, and PowerShell formats
- **Inline Secret Access** — Single secret retrieval via `--env-inline` flag
- **Global JSON Output** — Structured JSON output for all commands via `--json`
- **Cross-Platform** — Windows, Linux, and macOS support

## Installation

### Go Install

```bash
go install github.com/inovacc/profile@latest
```

### From Source

```bash
git clone https://github.com/inovacc/profile.git
cd profile
task build
```

## Quick Start

```bash
# Initialize the vault
profile vault init

# Create a profile
profile vault profile create myapp --default

# Store secrets
profile vault set API_KEY sk-abc123
profile vault set DB_PASSWORD s3cret -d "Production database"

# Retrieve a secret
profile vault get API_KEY

# List secrets
profile vault list
```

## Environment Variable Release

Release secrets as environment variables with password-gated, time-limited access:

```bash
# Set a password (min 8 characters)
profile vault env password set

# Release secrets (default: 30 minutes)
profile vault env release --ttl 1h

# Export as shell variables
eval $(profile vault env export --format export)       # bash/zsh
profile vault env export --format powershell | iex      # PowerShell

# Check release status
profile vault env status

# Revoke access
profile vault env revoke
```

## Import & Export

```bash
# Export secrets to JSON
profile vault export --format json > secrets.json

# Export as env file
profile vault export --format env > .env

# Import from file
profile vault import secrets.json
profile vault import .env --format env
```

## JSON Output

All commands support structured JSON output via the `--json` flag:

```bash
profile vault status --json
profile vault list --json
profile vault get API_KEY --json
```

## CLI Tools

```bash
# Display command tree
profile cmdtree

# Generate AI-readable documentation
profile aicontext
profile aicontext --compact
profile aicontext --category vault
```

## Command Reference

```
profile
├── vault
│   ├── init              Initialize the encrypted vault
│   ├── status            Show vault status
│   ├── set <key> <val>   Set an encrypted secret
│   ├── get <key>         Get a decrypted secret value
│   ├── delete <key>      Delete a secret
│   ├── list              List secrets in a profile
│   ├── export            Export secrets in plaintext
│   ├── import <file>     Import secrets from a file
│   ├── profile
│   │   ├── create        Create a new vault profile
│   │   ├── list          List all vault profiles
│   │   ├── delete        Delete a profile and its secrets
│   │   └── use           Set a profile as the default
│   └── env
│       ├── password
│       │   ├── set       Set or update the vault env password
│       │   └── reset     Remove the vault env password
│       ├── release       Release secrets for a time-limited period
│       ├── revoke        Revoke the active env release
│       ├── status        Show current env release status
│       └── export        Export released secrets as env variables
├── cmdtree               Display command tree
├── aicontext             Generate AI-readable documentation
└── completion            Shell completion scripts
```

## Security

- **AES-256-GCM** encryption for all stored secrets
- **HKDF-SHA256** key derivation bound to machine hardware identity
- **bcrypt** password hashing for env release gate
- **Time-limited sessions** with automatic expiry for env release
- Secrets are **never cached on disk** in plaintext
- Database is **non-portable** — only works on the originating machine

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
