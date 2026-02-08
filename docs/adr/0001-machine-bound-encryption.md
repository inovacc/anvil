# ADR-0001: Machine-Bound Encryption

## Status
Accepted

## Context
The vault needs to store secrets securely on disk. Secrets must be encrypted at rest and should not be portable to other machines without explicit export.

## Decision
Use AES-256-GCM with a master key that is sealed using a machine-derived key. The machine key is derived via HKDF-SHA256 from:
- The OS-specific machine identifier (Windows Registry GUID, Linux `/etc/machine-id`, macOS IOPlatformUUID)
- A random 32-byte salt
- A fixed info string `"profile-vault-v1"`

The sealed master key, nonce, and salt are stored in a singleton SQLite row.

## Consequences

### Positive
- Secrets are bound to the physical machine
- No password required for normal operations (machine identity is the authentication)
- Strong encryption with authenticated encryption (GCM)
- Key derivation is deterministic for the same machine

### Negative
- Secrets cannot be accessed on a different machine without explicit export/import
- Machine ID changes (hardware replacement, VM migration) will lock out the vault
- No recovery mechanism if machine identity is lost

## Alternatives Considered
- **Password-only encryption:** Simpler but requires password entry on every operation
- **TPM-based sealing:** More secure but platform-specific and complex
- **Cloud KMS:** Requires network access and cloud account
