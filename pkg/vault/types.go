package vault

import (
	"path/filepath"
	"time"

	"github.com/inovacc/anvil/internal/application"
)

// ProfileInfo represents vault profile metadata.
type ProfileInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	SecretCount int64     `json:"secret_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// SecretInfo represents secret metadata (without the decrypted value).
type SecretInfo struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SecretEntry represents a decrypted secret for export/import.
type SecretEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// Status represents the vault status information.
type Status struct {
	Initialized  bool      `json:"initialized"`
	DBPath       string    `json:"db_path"`
	ProfileCount int       `json:"profile_count"`
	SecretCount  int64     `json:"secret_count"`
	KeyVersion   int64     `json:"key_version"`
	SealMethod   string    `json:"seal_method"`
	PasswordSet  bool      `json:"password_set"`
	CreatedAt    time.Time `json:"created_at,omitzero"`
}

// ReleaseState represents the current state of an env release session.
type ReleaseState struct {
	Active      bool          `json:"active"`
	ProfileName string        `json:"profile_name,omitempty"`
	ExpiresAt   time.Time     `json:"expires_at,omitzero"`
	SessionID   string        `json:"session_id,omitempty"`
	Remaining   time.Duration `json:"remaining,omitempty"`
}

// EnvReleaseOptions configures an env release operation.
type EnvReleaseOptions struct {
	ProfileName string
	TTL         time.Duration
}

// AuditEntry represents a single audit log record.
type AuditEntry struct {
	Action      string    `json:"action"`
	ProfileName string    `json:"profile_name"`
	SecretKey   string    `json:"secret_key,omitempty"`
	Detail      string    `json:"detail,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// SecretVersion represents a historical version of a secret.
type SecretVersion struct {
	Version   int64     `json:"version"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

// Options configures vault behavior.
type Options struct {
	// DBPath overrides the default database path.
	DBPath string
}

// DefaultDBPath returns the default vault database path.
func DefaultDBPath() string {
	dir, err := application.GetApplicationDirectory()
	if err != nil {
		return ""
	}

	return filepath.Join(dir, dbFileName)
}
