# Installation ID Design

## Summary

Machine-bound deterministic installation ID derived from existing vault data. Queryable via `anvil id`, exposed in `pkg/vault` module.

## ID Derivation

```
InstallationID = hex(SHA-256(machine_id_hash || sealed_data))
```

- `machine_id_hash`: from `vault_sealed_key.machine_id_hash`
- `sealed_data`: from `vault_sealed_key.sealed_data`
- Result: 64-char lowercase hex string
- No extra storage — computed on the fly

## Changes

### `internal/crypto/installationid.go`

```go
func InstallationID(machineIDHash []byte, sealedData []byte) string
```

Pure function: SHA-256 of concatenated inputs, hex-encoded.

### `pkg/vault/iface.go`

New interface:

```go
type VaultIdentity interface {
    InstallationID() (string, error)
}
```

Compile-time check added.

### `pkg/vault/vault.go` (or new `identity.go`)

```go
func (v *Vault) InstallationID() (string, error)
```

Fetches sealed key from store, calls `crypto.InstallationID()`.

### `cmd/anvil/id.go`

Top-level `anvil id` command. Opens vault, prints ID, closes vault. Supports `--json`.

## Behavior

- Requires initialized vault (fails with `UserError` if not)
- Deterministic: same machine + same vault = same ID
- Changes on recovery to new machine (new machine ID + new sealed data)
- Does not leak machine ID or key material (double-hashed)

## Testing

- Unit test for `crypto.InstallationID()` — deterministic output
- Unit test: different inputs produce different IDs
- Integration test for `anvil id` command
