// Package vaultdb provides the SQLite database store for the encrypted vault.
// It wraps sqlc-generated queries with mutex-protected operations for profiles,
// secrets, sealed keys, and passwords.
package store
