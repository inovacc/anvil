// Package vault provides machine-bound encrypted secret storage organized by profiles.
//
// Secrets are encrypted with AES-256-GCM using a master key derived via HKDF-SHA256
// from the machine's hardware identity. The vault is non-portable by design — secrets
// encrypted on one machine cannot be decrypted on another.
//
// # Basic usage
//
//	// Initialize once per machine
//	vault.Init(nil)
//
//	// Open and use
//	v, _ := vault.Open(nil)
//	defer v.Close()
//
//	v.CreateProfile("myapp", "My application secrets", true)
//	v.Set("API_KEY", "sk-abc123", "myapp", "API key")
//
//	value, _ := v.Get("API_KEY", "myapp")
//
// # Interfaces
//
// The package provides interfaces for external consumers to program against
// and mock in tests:
//
//   - [VaultReader] — read-only operations (Get, List, Export, ListProfiles, Status)
//   - [VaultWriter] — extends VaultReader with write operations (Set, Delete, Import)
//   - [VaultEnv] — time-limited env release sessions (EnvRelease, EnvRevoke, EnvStatus)
//   - [VaultPassword] — password management (SetPassword, VerifyPassword, HasPassword)
package vault
