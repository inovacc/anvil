# ADR-0003: TPM 2.0 Hardware-Backed Master Key Sealing

## Status
Accepted

## Context
The vault seals the master key using software-only HKDF-SHA256 key derivation from the machine ID. The machine ID (Windows `MachineGuid` registry value, Linux `/etc/machine-id`, macOS serial number) is readable by any local process. This means master key protection depends solely on the secrecy of these values — insufficient against a local attacker with disk access.

Hardware security modules (TPM 2.0) are available on most modern Windows and Linux machines. TPM can seal data such that it can only be unsealed on the same physical hardware, with the key material never leaving the TPM chip.

## Decision
Use `github.com/inovacc/sealbox` to seal the master key to TPM 2.0 hardware when available, keeping the existing HKDF software path as a transparent fallback.

### Implementation
- **Init:** Try `sealbox.NewKeyManager()` → `km.SealKey(masterKey)`. On success, store JSON-serialized `SealedData` in `vault_sealed_key.sealed_data` with `seal_method = "tpm"`. On failure (no TPM, permission error), fall through to existing HKDF path with `seal_method = "software"`.
- **Open:** Switch on `seal_method` column to determine the unseal path. TPM path unmarshals `SealedData` and calls `km.UnsealKey()`. Software path uses existing HKDF + AES-GCM.
- **Close:** Zero master key via `sealbox.SecureZero()`.
- **Sentinel:** Use `sealbox.Encrypt`/`sealbox.Decrypt` for packed nonce||ciphertext format (binary-compatible with existing sentinel files).
- **Migration:** `seal_method TEXT NOT NULL DEFAULT 'software'` column added to `vault_sealed_key`. Idempotent `ALTER TABLE` in `store.Open()` handles pre-migration databases.

### Platform Support
| Platform | TPM Support | Fallback |
|----------|-------------|----------|
| Linux | `/dev/tpmrm0` (user in `tss` group) | Software HKDF |
| Windows | TPM Base Services (TBS) | Software HKDF |
| macOS | Not yet (Secure Enclave planned) | Software HKDF |

## Consequences

### Positive
- Master key is hardware-bound on TPM-equipped machines; cannot be extracted even with full disk access
- Transparent fallback means existing vaults and non-TPM machines continue working without changes
- `seal_method` discriminator makes the vault self-describing
- Memory zeroing prevents key material from lingering in process memory
- Sentinel packed encryption simplifies sentinel code (no manual nonce management)

### Negative
- New dependency on `github.com/inovacc/sealbox` (which pulls in `google/go-tpm`)
- TPM operations are slower (~3s for seal/unseal vs microseconds for HKDF)
- TPM requires platform-specific permissions (Linux: `tss` group; Windows: admin for some operations)
- Vault sealed with TPM on one machine cannot be migrated, even to another machine with TPM (by design)

### Neutral
- Existing software-sealed vaults are unaffected; migration adds `seal_method = 'software'` as default
- No change to the secret-level encryption (AES-256-GCM with master key remains the same)
