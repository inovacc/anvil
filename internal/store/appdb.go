package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// AppSecret represents a secret stored in an app database.
type AppSecret struct {
	ID             int64
	Key            string
	EncryptedValue []byte
	Nonce          []byte
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// AppStore provides operations on an isolated app database.
type AppStore struct {
	db *sql.DB
	mu sync.RWMutex
}

// OpenApp opens or creates an app database at the given path.
func OpenApp(dbPath string) (*AppStore, error) {
	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open app database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec(AppSchema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create app schema: %w", err)
	}

	return &AppStore{db: db}, nil
}

// Close closes the app database.
func (s *AppStore) Close() error {
	return s.db.Close()
}

// Set creates or updates a secret.
func (s *AppStore) Set(key string, encryptedValue, nonce []byte, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO app_secrets (key, encrypted_value, nonce, description)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET
		   encrypted_value = excluded.encrypted_value,
		   nonce = excluded.nonce,
		   description = excluded.description,
		   updated_at = CURRENT_TIMESTAMP`,
		key, encryptedValue, nonce, description)

	return err
}

// Get retrieves a secret by key.
func (s *AppStore) Get(key string) (AppSecret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sec AppSecret

	err := s.db.QueryRowContext(context.Background(),
		`SELECT id, key, encrypted_value, nonce, description, created_at, updated_at
		 FROM app_secrets WHERE key = ? LIMIT 1`, key).Scan(
		&sec.ID, &sec.Key, &sec.EncryptedValue, &sec.Nonce,
		&sec.Description, &sec.CreatedAt, &sec.UpdatedAt)

	return sec, err
}

// Delete removes a secret by key.
func (s *AppStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	result, err := s.db.ExecContext(context.Background(),
		`DELETE FROM app_secrets WHERE key = ?`, key)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List returns all secrets.
func (s *AppStore) List() ([]AppSecret, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(context.Background(),
		`SELECT id, key, encrypted_value, nonce, description, created_at, updated_at
		 FROM app_secrets ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var secrets []AppSecret

	for rows.Next() {
		var sec AppSecret

		if err := rows.Scan(&sec.ID, &sec.Key, &sec.EncryptedValue, &sec.Nonce,
			&sec.Description, &sec.CreatedAt, &sec.UpdatedAt); err != nil {
			return nil, err
		}

		secrets = append(secrets, sec)
	}

	return secrets, rows.Err()
}

// Count returns the number of secrets.
func (s *AppStore) Count() (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int64

	err := s.db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM app_secrets`).Scan(&count)

	return count, err
}

// ListAll returns all secrets (used for key rotation re-encryption).
func (s *AppStore) ListAll() ([]AppSecret, error) {
	return s.List()
}

// UpdateSecret updates the encrypted value and nonce for a secret by ID.
func (s *AppStore) UpdateSecret(id int64, encryptedValue, nonce []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(context.Background(),
		`UPDATE app_secrets SET encrypted_value = ?, nonce = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		encryptedValue, nonce, id)

	return err
}
