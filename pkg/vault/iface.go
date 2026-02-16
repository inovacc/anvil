package vault

import "time"

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

// VaultAudit provides audit log operations.
type VaultAudit interface {
	AuditLog(limit int64) ([]AuditEntry, error)
	AuditLogByProfile(profileName string, limit int64) ([]AuditEntry, error)
	PurgeAuditLog(before time.Time) error
}

// VaultVersioning provides secret version history and rollback operations.
type VaultVersioning interface {
	SecretHistory(key, profileName string) ([]SecretVersion, error)
	SecretRollback(key, profileName string, version int64) error
}

// VaultTemplate provides secret template operations.
type VaultTemplate interface {
	CreateTemplate(def *TemplateDefinition) error
	GetTemplate(name string) (*TemplateDefinition, error)
	ListTemplates() ([]TemplateInfo, error)
	DeleteTemplate(name string) error
	ApplyTemplate(name, profileName string, vars map[string]string) error
}

// VaultBackup provides backup, restore, and sharing operations.
type VaultBackup interface {
	Backup(password string) ([]byte, error)
	Restore(encrypted []byte, password string) error
	ShareExport(profileName, passphrase string) ([]byte, error)
	ShareImport(encrypted []byte, passphrase, targetProfile string) (*SharedExport, error)
}

// VaultScoped provides isolated, single-profile access to the vault.
// External apps use this interface — they can only read/write secrets in their own profile.
type VaultScoped interface {
	Get(key string) (string, error)
	Set(key, value, description string) error
	Delete(key string) error
	List() ([]SecretInfo, error)
	Export() ([]SecretEntry, error)
	Import(entries []SecretEntry) error
	ProfileInfo() ProfileInfo
	Close() error
}

// Compile-time interface satisfaction checks.
var (
	_ VaultReader      = (*Vault)(nil)
	_ VaultWriter      = (*Vault)(nil)
	_ VaultEnv         = (*Vault)(nil)
	_ VaultPassword    = (*Vault)(nil)
	_ VaultKeyRotation = (*Vault)(nil)
	_ VaultAudit       = (*Vault)(nil)
	_ VaultVersioning  = (*Vault)(nil)
	_ VaultTemplate    = (*Vault)(nil)
	_ VaultBackup      = (*Vault)(nil)
)
