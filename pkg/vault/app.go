package vault

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/inovacc/anvil/internal/crypto"
	"github.com/inovacc/anvil/internal/store"
	"github.com/inovacc/anvil/internal/store/sqlc"
)

// AppVault provides encrypted secret storage scoped to a single app database.
type AppVault struct {
	store       *store.AppStore
	masterKey   []byte
	appUUID     string
	appName     string
	parentStore *store.Store
}

// RegisterApp creates a new isolated app database and registers it in the main vault.
func (v *Vault) RegisterApp(name, description, serviceID string) (*AppInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	_, err := v.store.GetAppByName(name)
	if err == nil {
		return nil, &UserError{
			Message: fmt.Sprintf("App %q already exists.", name),
			Hint:    "Choose a different name or remove the existing app first.",
		}
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("check app: %w", err)
	}

	appUUID := uuid.New().String()
	appsDir := filepath.Join(filepath.Dir(v.dbPath), "apps")

	if err := os.MkdirAll(appsDir, 0o700); err != nil {
		return nil, fmt.Errorf("create apps directory: %w", err)
	}

	dbPath := filepath.Join(appsDir, appUUID+".db")

	appStore, err := store.OpenApp(dbPath)
	if err != nil {
		return nil, fmt.Errorf("create app database: %w", err)
	}

	_ = appStore.Close()

	if err := v.store.RegisterApp(appUUID, name, description, serviceID, dbPath); err != nil {
		_ = os.Remove(dbPath)
		return nil, fmt.Errorf("register app: %w", err)
	}

	v.logAudit("app.register", "", "", fmt.Sprintf("name=%s uuid=%s", name, appUUID))

	return &AppInfo{
		UUID:        appUUID,
		Name:        name,
		Description: description,
		ServiceID:   serviceID,
		DBPath:      dbPath,
		Status:      "active",
		CreatedAt:   time.Now(),
	}, nil
}

// ListApps returns all registered apps.
func (v *Vault) ListApps() ([]AppInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	rows, err := v.store.ListApps()
	if err != nil {
		return nil, fmt.Errorf("list apps: %w", err)
	}

	apps := make([]AppInfo, 0, len(rows))

	for _, row := range rows {
		apps = append(apps, appInfoFromRow(row))
	}

	return apps, nil
}

// GetApp retrieves an app by name or UUID.
func (v *Vault) GetApp(nameOrUUID string) (*AppInfo, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	row, err := v.resolveApp(nameOrUUID)
	if err != nil {
		return nil, err
	}

	info := appInfoFromRow(*row)

	return &info, nil
}

// RemoveApp removes an app and its database file.
func (v *Vault) RemoveApp(nameOrUUID string) error {
	if err := v.checkSealed(); err != nil {
		return err
	}

	row, err := v.resolveApp(nameOrUUID)
	if err != nil {
		return err
	}

	if err := os.Remove(row.DbPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove app database: %w", err)
	}

	if err := v.store.DeleteApp(row.Uuid); err != nil {
		return fmt.Errorf("delete app registry: %w", err)
	}

	v.logAudit("app.remove", "", "", fmt.Sprintf("name=%s uuid=%s", row.Name, row.Uuid))

	return nil
}

// DisableApp sets an app's status to disabled.
func (v *Vault) DisableApp(nameOrUUID string) error {
	if err := v.checkSealed(); err != nil {
		return err
	}

	row, err := v.resolveApp(nameOrUUID)
	if err != nil {
		return err
	}

	if err := v.store.UpdateAppStatus(row.Uuid, "disabled"); err != nil {
		return fmt.Errorf("disable app: %w", err)
	}

	v.logAudit("app.disable", "", "", fmt.Sprintf("name=%s", row.Name))

	return nil
}

// EnableApp sets an app's status to active.
func (v *Vault) EnableApp(nameOrUUID string) error {
	if err := v.checkSealed(); err != nil {
		return err
	}

	row, err := v.resolveApp(nameOrUUID)
	if err != nil {
		return err
	}

	if err := v.store.UpdateAppStatus(row.Uuid, "active"); err != nil {
		return fmt.Errorf("enable app: %w", err)
	}

	v.logAudit("app.enable", "", "", fmt.Sprintf("name=%s", row.Name))

	return nil
}

// OpenApp opens an app's database and returns an AppVault for secret operations.
func (v *Vault) OpenApp(nameOrUUID string) (*AppVault, error) {
	if err := v.checkSealed(); err != nil {
		return nil, err
	}

	row, err := v.resolveApp(nameOrUUID)
	if err != nil {
		return nil, err
	}

	if row.Status != "active" {
		return nil, &UserError{
			Message: fmt.Sprintf("App %q is %s.", row.Name, row.Status),
			Hint:    "Enable the app first with: anvil app enable " + row.Name,
		}
	}

	appStore, err := store.OpenApp(row.DbPath)
	if err != nil {
		return nil, fmt.Errorf("open app database: %w", err)
	}

	_ = v.store.UpdateAppLastAccessed(row.Uuid)

	return &AppVault{
		store:       appStore,
		masterKey:   v.masterKey,
		appUUID:     row.Uuid,
		appName:     row.Name,
		parentStore: v.store,
	}, nil
}

// resolveApp looks up an app by name or UUID.
func (v *Vault) resolveApp(nameOrUUID string) (*sqlc.VaultApp, error) {
	row, err := v.store.GetAppByUUID(nameOrUUID)
	if err == nil {
		return &row, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("get app: %w", err)
	}

	row, err = v.store.GetAppByName(nameOrUUID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &UserError{
			Message: fmt.Sprintf("App %q not found.", nameOrUUID),
			Hint:    "Use 'anvil app list' to see registered apps.",
		}
	}

	if err != nil {
		return nil, fmt.Errorf("get app: %w", err)
	}

	return &row, nil
}

// appInfoFromRow converts a sqlc VaultApp row to an AppInfo.
func appInfoFromRow(row sqlc.VaultApp) AppInfo {
	info := AppInfo{
		UUID:      row.Uuid,
		Name:      row.Name,
		DBPath:    row.DbPath,
		Status:    row.Status,
		CreatedAt: row.CreatedAt,
	}

	if row.Description != nil {
		info.Description = *row.Description
	}

	if row.ServiceID != nil {
		info.ServiceID = *row.ServiceID
	}

	if row.SecretCount != nil {
		info.SecretCount = *row.SecretCount
	}

	if row.LastAccessedAt != nil {
		info.LastAccessedAt = *row.LastAccessedAt
	}

	return info
}

// === AppVault operations ===

// Set encrypts and stores a secret in the app database.
func (av *AppVault) Set(key, value, description string) error {
	ciphertext, nonce, err := crypto.Encrypt(av.masterKey, []byte(value))
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := av.store.Set(key, ciphertext, nonce, description); err != nil {
		return err
	}

	av.updateSecretCount()

	return nil
}

// Get decrypts and returns a secret value.
func (av *AppVault) Get(key string) (string, error) {
	sec, err := av.store.Get(key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrSecretNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}

	plaintext, err := crypto.Decrypt(av.masterKey, sec.EncryptedValue, sec.Nonce)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Delete removes a secret.
func (av *AppVault) Delete(key string) error {
	if err := av.store.Delete(key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSecretNotFound
		}

		return fmt.Errorf("delete secret: %w", err)
	}

	av.updateSecretCount()

	return nil
}

// List returns metadata for all secrets.
func (av *AppVault) List() ([]SecretInfo, error) {
	secrets, err := av.store.List()
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	infos := make([]SecretInfo, 0, len(secrets))

	for _, sec := range secrets {
		infos = append(infos, SecretInfo{
			Key:         sec.Key,
			Description: sec.Description,
			CreatedAt:   sec.CreatedAt,
			UpdatedAt:   sec.UpdatedAt,
		})
	}

	return infos, nil
}

// Export decrypts and returns all secrets.
func (av *AppVault) Export() ([]SecretEntry, error) {
	secrets, err := av.store.List()
	if err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}

	entries := make([]SecretEntry, 0, len(secrets))

	for _, sec := range secrets {
		plaintext, err := crypto.Decrypt(av.masterKey, sec.EncryptedValue, sec.Nonce)
		if err != nil {
			return nil, fmt.Errorf("decrypt %q: %w", sec.Key, err)
		}

		entries = append(entries, SecretEntry{
			Key:         sec.Key,
			Value:       string(plaintext),
			Description: sec.Description,
		})
	}

	return entries, nil
}

// Import encrypts and stores multiple secrets.
func (av *AppVault) Import(entries []SecretEntry) error {
	for _, entry := range entries {
		ciphertext, nonce, err := crypto.Encrypt(av.masterKey, []byte(entry.Value))
		if err != nil {
			return fmt.Errorf("encrypt %q: %w", entry.Key, err)
		}

		if err := av.store.Set(entry.Key, ciphertext, nonce, entry.Description); err != nil {
			return fmt.Errorf("save %q: %w", entry.Key, err)
		}
	}

	av.updateSecretCount()

	return nil
}

// Close closes the app database.
func (av *AppVault) Close() error {
	if av == nil {
		return nil
	}

	return av.store.Close()
}

// updateSecretCount updates the cached count in the main vault registry.
func (av *AppVault) updateSecretCount() {
	count, err := av.store.Count()
	if err != nil {
		return
	}

	_ = av.parentStore.UpdateAppSecretCount(av.appUUID, count)
}
