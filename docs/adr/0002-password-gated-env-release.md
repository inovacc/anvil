# ADR-0002: Password-Gated Environment Variable Release

## Status
Accepted

## Context
External applications need to consume vault secrets as environment variables. The vault's machine-binding provides baseline security, but exposing secrets to the environment requires an additional authentication gate and time-limited access to reduce the window of exposure.

## Decision
Implement a password-gated release mechanism with:
- **Password storage:** bcrypt hash stored in the vault database (cost >= 10)
- **Sentinel file:** Encrypted session state stored in `~/.cache/profile/` using AES-256-GCM with the vault master key
- **TTL-based access:** Secrets are released for a configurable duration (1 minute to 24 hours)
- **Auto-expiry:** Expired sessions are automatically revoked on check
- **Inline access:** `profile --env-inline KEY` for shell command substitution

The sentinel file format is `nonce (12 bytes) || AES-256-GCM ciphertext(JSON payload)`, making it opaque without vault access.

## Consequences

### Positive
- Additional authentication layer beyond machine binding
- Time-limited exposure window for secrets
- Multiple output formats (env, export, powershell) for cross-platform support
- Secrets are never cached on disk in plaintext
- Sentinel file is encrypted and machine-bound

### Negative
- Requires password entry for each release session
- File-based sentinel can be deleted (though secrets remain encrypted in vault)
- No remote revocation capability

## Alternatives Considered
- **Unix socket server:** More secure but complex, requires daemon management
- **Named pipe:** Platform-specific, Windows vs Unix differences
- **Environment variable injection via process:** Requires elevated privileges
