package vault

// VaultReader provides read-only access to a vault.
type VaultReader interface {
	Get(key, profileName string) (string, error)
	List(profileName string) ([]SecretInfo, error)
	Export(profileName string) ([]SecretEntry, error)
	ListProfiles() ([]ProfileInfo, error)
	Status() (*Status, error)
	DBPath() string
	Close() error
}

// VaultWriter extends VaultReader with write operations.
type VaultWriter interface {
	VaultReader
	Set(key, value, profileName, description string) error
	Delete(key, profileName string) error
	Import(entries []SecretEntry, profileName string) error
	CreateProfile(name, description string, isDefault bool) error
	DeleteProfile(name string) error
	UseProfile(name string) error
}

// VaultEnv provides env release operations.
type VaultEnv interface {
	EnvRelease(password string, opts *EnvReleaseOptions) (*ReleaseState, error)
	EnvRevoke() error
	EnvStatus() (*ReleaseState, error)
	EnvExport(profileName string) ([]SecretEntry, error)
	EnvInlineGet(key, profileName string) (string, error)
}

// VaultPassword provides password management operations.
type VaultPassword interface {
	SetPassword(password string) error
	VerifyPassword(password string) error
	HasPassword() (bool, error)
	DeletePassword() error
}

// VaultKeyRotation provides master key rotation operations.
type VaultKeyRotation interface {
	RotateKey(password string) error
}

// Compile-time interface satisfaction checks.
var (
	_ VaultReader   = (*Vault)(nil)
	_ VaultWriter   = (*Vault)(nil)
	_ VaultEnv      = (*Vault)(nil)
	_ VaultPassword    = (*Vault)(nil)
	_ VaultKeyRotation = (*Vault)(nil)
)
