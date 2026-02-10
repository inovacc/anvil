package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/inovacc/profile/internal/store/sqlc"
	_ "modernc.org/sqlite"
)

// Store provides vault database operations.
type Store struct {
	db      *sql.DB
	queries *sqlc.Queries
	mu      sync.RWMutex
}

// Open opens or creates the vault database at the given path.
func Open(dbPath string) (*Store, error) {
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(Schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return &Store{
		db:      db,
		queries: sqlc.New(db),
	}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Ping verifies database connectivity.
func (s *Store) Ping() error {
	return s.db.Ping()
}

// === Profile operations ===

// CreateProfile creates a new vault profile.
func (s *Store) CreateProfile(name, description string, isDefault bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	if isDefault {
		if err := s.queries.ClearAllDefaultProfiles(ctx); err != nil {
			return err
		}
	}

	def := int64(0)
	if isDefault {
		def = 1
	}

	return s.queries.CreateProfile(ctx, sqlc.CreateProfileParams{
		Name:        name,
		Description: &description,
		IsDefault:   &def,
	})
}

// GetProfile retrieves a profile by name.
func (s *Store) GetProfile(name string) (sqlc.VaultProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.GetProfileByName(context.Background(), name)
}

// GetDefaultProfile retrieves the default profile.
func (s *Store) GetDefaultProfile() (sqlc.VaultProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.GetDefaultProfile(context.Background())
}

// ListProfiles returns all vault profiles.
func (s *Store) ListProfiles() ([]sqlc.VaultProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.ListProfiles(context.Background())
}

// ProfileExists checks if a profile exists.
func (s *Store) ProfileExists(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count, err := s.queries.ProfileExists(context.Background(), name)

	return count > 0, err
}

// DeleteProfile deletes a profile and all its secrets (FK cascade).
func (s *Store) DeleteProfile(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.DeleteProfile(context.Background(), name)
}

// SetDefaultProfile sets a profile as the default.
func (s *Store) SetDefaultProfile(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	if err := s.queries.ClearAllDefaultProfiles(ctx); err != nil {
		return err
	}

	return s.queries.SetProfileDefault(ctx, name)
}

// === Secret operations ===

// UpsertSecret creates or updates an encrypted secret.
func (s *Store) UpsertSecret(profileName, key string, encryptedValue, nonce []byte, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.UpsertSecret(context.Background(), sqlc.UpsertSecretParams{
		ProfileName:    profileName,
		Key:            key,
		EncryptedValue: encryptedValue,
		Nonce:          nonce,
		Description:    &description,
	})
}

// GetSecret retrieves an encrypted secret.
func (s *Store) GetSecret(profileName, key string) (sqlc.VaultSecret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.GetSecret(context.Background(), sqlc.GetSecretParams{
		ProfileName: profileName,
		Key:         key,
	})
}

// ListSecrets returns all secrets for a profile.
func (s *Store) ListSecrets(profileName string) ([]sqlc.VaultSecret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.ListSecrets(context.Background(), profileName)
}

// DeleteSecret deletes a secret.
func (s *Store) DeleteSecret(profileName, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.DeleteSecret(context.Background(), sqlc.DeleteSecretParams{
		ProfileName: profileName,
		Key:         key,
	})
}

// SecretExists checks if a secret exists.
func (s *Store) SecretExists(profileName, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count, err := s.queries.SecretExists(context.Background(), sqlc.SecretExistsParams{
		ProfileName: profileName,
		Key:         key,
	})

	return count > 0, err
}

// CountSecrets returns the number of secrets for a profile.
func (s *Store) CountSecrets(profileName string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.CountSecrets(context.Background(), profileName)
}

// === Sealed key operations ===

// GetSealedKey retrieves the sealed master key.
func (s *Store) GetSealedKey() (sqlc.VaultSealedKey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.GetSealedKey(context.Background())
}

// HasSealedKey checks if a sealed key exists.
func (s *Store) HasSealedKey() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count, err := s.queries.HasSealedKey(context.Background())

	return count > 0, err
}

// UpsertSealedKey saves the sealed master key.
func (s *Store) UpsertSealedKey(sealedData, nonce, keySalt, machineIDHash []byte, version int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.UpsertSealedKey(context.Background(), sqlc.UpsertSealedKeyParams{
		SealedData:    sealedData,
		Nonce:         nonce,
		KeySalt:       keySalt,
		Version:       &version,
		MachineIDHash: machineIDHash,
	})
}

// DeleteSealedKey deletes the sealed master key.
func (s *Store) DeleteSealedKey() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.DeleteSealedKey(context.Background())
}

// === Password operations ===

// GetPassword retrieves the vault password hash.
func (s *Store) GetPassword() (sqlc.VaultPassword, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.queries.GetPassword(context.Background())
}

// HasPassword checks if a password has been set.
func (s *Store) HasPassword() (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count, err := s.queries.HasPassword(context.Background())

	return count > 0, err
}

// UpsertPassword saves the vault password hash.
func (s *Store) UpsertPassword(passwordHash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.UpsertPassword(context.Background(), passwordHash)
}

// DeletePassword deletes the vault password.
func (s *Store) DeletePassword() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.queries.DeletePassword(context.Background())
}
