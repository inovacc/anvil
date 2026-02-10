package vault

import (
	"time"

	"github.com/inovacc/profile/internal/application"
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

// EnvReleaseOptions configures an env release operation.
type EnvReleaseOptions struct {
	ProfileName string
	TTL         time.Duration
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

	return dir
}
