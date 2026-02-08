// Package crypto provides AES-256-GCM encryption and machine-bound key management.
//
// The master key is sealed using HKDF-SHA256 with the machine's hardware identity
// as input key material, ensuring secrets are non-portable across machines.
// Platform-specific machine ID detection supports Windows, Linux, and macOS.
package crypto
