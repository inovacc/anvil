package vault

import "time"

// Options configures vault operations.
type Options struct {
	DBPath string
}

// ProfileInfo contains profile metadata.
type ProfileInfo struct {
	Name        string    `json:"name"`
	UUID        string    `json:"uuid"`
	Description string    `json:"description"`
	IsDefault   bool      `json:"is_default"`
	SecretCount int64     `json:"secret_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// SecretInfo contains secret metadata (without the value).
type SecretInfo struct {
	Key         string    `json:"key"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SecretEntry contains a decrypted secret key-value pair.
type SecretEntry struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

// Status contains vault status information.
type Status struct {
	Initialized  bool      `json:"initialized"`
	DBPath       string    `json:"db_path"`
	ProfileCount int       `json:"profile_count"`
	SecretCount  int64     `json:"secret_count"`
	KeyVersion   int64     `json:"key_version"`
	SealMethod   string    `json:"seal_method"`
	PasswordSet  bool      `json:"password_set"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuditEntry represents an audit log entry.
type AuditEntry struct {
	Action      string    `json:"action"`
	ProfileName string    `json:"profile_name"`
	SecretKey   string    `json:"secret_key,omitempty"`
	Detail      string    `json:"detail,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// SecretVersion represents metadata for an archived secret version.
// The value is never exposed — only rollback can restore it.
type SecretVersion struct {
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}

// ReleaseState represents an env release session.
type ReleaseState struct {
	Active      bool          `json:"active"`
	ProfileName string        `json:"profile_name"`
	ExpiresAt   time.Time     `json:"expires_at"`
	SessionID   string        `json:"session_id"`
	Remaining   time.Duration `json:"remaining"`
}

// AppInfo contains metadata about a registered app.
type AppInfo struct {
	UUID           string    `json:"uuid"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	ServiceID      string    `json:"service_id"`
	DBPath         string    `json:"db_path"`
	Status         string    `json:"status"`
	SecretCount    int64     `json:"secret_count"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

// KeyAlgorithm represents a supported asymmetric key algorithm.
type KeyAlgorithm = string

const (
	AlgorithmEd25519   KeyAlgorithm = "ed25519"
	AlgorithmECDSAP256 KeyAlgorithm = "ecdsa-p256"
)

// KeyInfo contains asymmetric key metadata.
type KeyInfo struct {
	Name        string    `json:"name"`
	Algorithm   string    `json:"algorithm"`
	Fingerprint string    `json:"fingerprint"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// SignResult contains the output of a sign operation.
type SignResult struct {
	Signature string `json:"signature"`
	KeyName   string `json:"key_name"`
	Algorithm string `json:"algorithm"`
}

// VerifyResult contains the output of a verify operation.
type VerifyResult struct {
	Valid     bool   `json:"valid"`
	KeyName   string `json:"key_name"`
	Algorithm string `json:"algorithm"`
}

// EnvReleaseOptions configures env release sessions.
type EnvReleaseOptions struct {
	ProfileName string
	TTL         time.Duration
}
