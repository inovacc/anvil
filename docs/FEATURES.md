# Features

## Completed

### Machine-Bound Encryption

- **Status:** Completed (v0.1.0)
- AES-256-GCM encryption for all stored secrets
- HKDF-SHA256 key derivation bound to machine hardware identity
- Sealed master key stored in database, non-portable by design

### Profile Management

- **Status:** Completed (v0.1.0)
- Create, list, delete named profiles
- Set a default profile for convenience
- Secrets organized per profile

### Secret CRUD

- **Status:** Completed (v0.1.0)
- Set, get, delete, list encrypted secrets
- Optional descriptions on secrets
- Secret existence checks

### Import & Export

- **Status:** Completed (v0.1.0)
- Export secrets as JSON, env, bash export, or PowerShell format
- Import from JSON or env files

### Password-Gated Environment Release

- **Status:** Completed (v0.2.0)
- bcrypt password hashing for vault access gate
- Time-limited secret release sessions (configurable TTL)
- File-based sentinel state management
- Auto-expiry and manual revoke
- Inline secret access via `--env-inline` flag

### Global JSON Output

- **Status:** Completed (v0.2.0)
- All commands support `--json` flag for structured output
- Dual output mode: human-readable text or machine-parseable JSON

### CLI Tooling

- **Status:** Completed (v0.2.0)
- Command tree visualization (`cmdtree`)
- AI-readable documentation generator (`aicontext`)
- Shell completion scripts

### TPM 2.0 Hardware-Backed Key Sealing

- **Status:** Completed (v0.3.0)
- Master key sealed to TPM 2.0 hardware via sealbox when available
- Transparent software fallback (HKDF) for machines without TPM
- `seal_method` column discriminates TPM vs software unseal path
- Platform support: Windows (TBS), Linux (`/dev/tpmrm0`); macOS falls back to software
- `vault status` displays current seal method

### Memory Safety

- **Status:** Completed (v0.3.0)
- Master key zeroed on vault close and after init via `sealbox.SecureZero`
- Master key only lives in memory during the vault lifecycle

### Sentinel Packed Encryption

- **Status:** Completed (v0.3.0)
- Sentinel files use `sealbox.Encrypt`/`sealbox.Decrypt` (packed nonce||ciphertext)
- Binary-compatible with the original manual nonce packing format

### CI/CD Release Automation

- **Status:** Completed (v0.3.0)
- GitHub Actions release workflow triggers goreleaser on tag push
- Test workflow runs lint, tests, and vulncheck on PRs
- Build workflow validates Linux and Windows builds on develop branch

### User-Friendly Error Handling

- **Status:** Completed (v0.3.x)
- `UserError` type with Message and optional Hint for all vault errors
- `handleError` formats errors for both text and JSON (`--json`) output
- Cobra `SilenceErrors`/`SilenceUsage` suppresses usage dump on errors
- Actionable hints guide users to the correct command

### Public Go API with Interfaces

- **Status:** Completed (v0.3.x)
- Clean `pkg/vault` module boundary — no `internal/` types in public signatures
- `VaultReader` interface for read-only vault consumers
- `VaultWriter` interface extending VaultReader with write operations
- `VaultEnv` interface for env release session management
- `VaultPassword` interface for password operations
- Compile-time interface satisfaction checks (`var _ VaultReader = (*Vault)(nil)`)
- Vault-owned `ReleaseState` type (decoupled from internal sentinel package)
- Removed `MasterKey()` method to prevent raw key exposure

### Master Key Rotation

- **Status:** Completed (v0.4.0)
- `vault rotate-key` with transactional re-encryption of all secrets and versions
- Password verification required; revokes active sentinel session

### Audit Logging

- **Status:** Completed (v0.4.0)
- `vault audit` command with profile filtering and purge
- Tracks all CRUD, rotation, env, backup, and share operations

### Secret Versioning

- **Status:** Completed (v0.4.0)
- `vault history <key>` and `vault rollback <key> <version>`
- Automatic archival on overwrite; delete removes all versions

### Shell Integration Helpers

- **Status:** Completed (v0.5.0)
- Native auto-completion for bash, zsh, fish, powershell
- Dynamic completion for profile names and secret keys

### Docker Secrets Bridge

- **Status:** Completed (v0.5.0)
- `vault docker export/clean/compose` commands
- Writes one file per secret; generates Docker Compose YAML snippet

### Backup & Restore

- **Status:** Completed (v0.4.0)
- Password-encrypted full vault backup (profiles, secrets, versions, password hash)
- `vault backup` and `vault restore` commands

### Encrypted Secret Sharing

- **Status:** Completed (v0.4.0)
- `vault share export/import` with passphrase-based encryption
- Cross-machine secret transfer

### Secret Templates

- **Status:** Completed (v0.5.0)
- `vault template create/list/show/delete/apply` commands
- 5 built-in templates: postgres, mysql, redis, aws, github-token
- Variable interpolation via Go `text/template`

### Plugin System

- **Status:** Completed (v0.5.0)
- Event hooks for secret lifecycle (pre/post set, get, delete, rotate, init)
- Pre-hooks can block operations by returning `{"allow":false}` via JSON stdout
- Secret provider plugins for external sources (JSON stdin/stdout protocol)
- CLI management: `vault plugin hook-add/hook-remove/provider-add/provider-remove/list`
- Config stored as `plugins.json` alongside vault DB (no schema migration)
- Hooks auto-fire on `Set`, `Get`, `Delete` operations via integrated `PluginManager`

### Interactive TUI

- **Status:** Completed (v0.7.0)
- `vault tui` command with dashboard, profile browser, and secrets table view
- Built with bubbletea/lipgloss/bubbles (table component for secrets with Key/Description/Created/Updated columns)
- Keyboard-driven navigation: dashboard → profiles → secrets → secret form

### Per-App Isolated Vaults

- **Status:** Completed (v0.8.0)
- Register external apps with dedicated vault databases
- UUID-based app identification, scoped secret access
- `vault app register/list/remove/disable/enable` commands

### BIP-39 Mnemonic Recovery

- **Status:** Completed (v0.8.0)
- 24-word BIP-39 recovery phrase generated during vault init
- Allows vault master key recovery on a new machine if TPM dies or machine changes
- `recover` command re-seals master key to current machine from mnemonic
- `recovery-phrase` command displays the 24 words (requires password verification)
- Only SHA-256 hash of mnemonic stored in DB — mnemonic never persisted

### Machine-Bound Installation ID

- **Status:** Completed (v0.8.x)
- Deterministic `SHA-256(machine_id_hash || sealed_data)` identifier
- Queryable via `anvil id` (supports `--json`)
- Exposed in `pkg/vault` via `VaultIdentity` interface
- No extra storage — computed from existing `vault_sealed_key` data
- Changes on vault recovery to new machine (new machine ID + new sealed data)

### Asymmetric Key Management

- **Status:** Completed (v0.9.0)
- Ed25519 (default) and ECDSA P-256 key pair generation and storage
- Private keys encrypted with vault master key (AES-256-GCM) in `vault_keys` table
- Public keys stored unencrypted for verification without decryption
- `key generate/list/delete/export/import` commands
- PEM export/import using PKCS8 (private) and PKIX (public) standard formats
- SHA-256 fingerprint (first 8 bytes hex) for key identification

### Digital Signing & Verification

- **Status:** Completed (v0.9.0)
- `sign --key <name> (--file | --string)` — outputs base64-encoded signature
- `verify --key <name> (--file | --string) (--signature | --signature-file)` — exits 0 (valid) or 1 (invalid)
- Supports both Ed25519 and ECDSA P-256 algorithms
- File output via `-o` flag for signatures
- `VaultKeyManagement` and `VaultSigner` interfaces for programmatic access

### MCP Server Integration

- **Status:** Completed (v0.10.0)
- MCP server exposing vault operations as 17 tools via Go SDK (`github.com/modelcontextprotocol/go-sdk/mcp`)
- Tools: `secret_get`, `secret_set`, `secret_delete`, `secret_list`, `profile_list`, `profile_create`, `profile_delete`, `key_generate`, `key_list`, `key_delete`, `key_export`, `sign`, `verify`, `audit_log`, `vault_status`, `installation_id`
- Resource: `anvil://status` — vault status as JSON
- Stdio transport for CLI integration (`anvil mcp serve`)
- In-memory transport for testing

## Proposed

_No proposed features at this time._
