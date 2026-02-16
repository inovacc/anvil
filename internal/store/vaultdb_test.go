package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/inovacc/anvil/internal/store/sqlc"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q): %v", dbPath, err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestOpenAndPing(t *testing.T) {
	s := openTestStore(t)
	if err := s.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestProfileCRUD(t *testing.T) {
	s := openTestStore(t)

	t.Run("create and get", func(t *testing.T) {
		if err := s.CreateProfile("dev", "development", true, ""); err != nil {
			t.Fatalf("CreateProfile: %v", err)
		}
		p, err := s.GetProfile("dev")
		if err != nil {
			t.Fatalf("GetProfile: %v", err)
		}
		if p.Name != "dev" {
			t.Errorf("got name %q, want %q", p.Name, "dev")
		}
		if p.Description == nil || *p.Description != "development" {
			t.Errorf("got description %v, want %q", p.Description, "development")
		}
		if p.IsDefault == nil || *p.IsDefault != 1 {
			t.Errorf("expected is_default=1, got %v", p.IsDefault)
		}
	})

	t.Run("get default profile", func(t *testing.T) {
		p, err := s.GetDefaultProfile()
		if err != nil {
			t.Fatalf("GetDefaultProfile: %v", err)
		}
		if p.Name != "dev" {
			t.Errorf("got %q, want %q", p.Name, "dev")
		}
	})

	t.Run("list profiles", func(t *testing.T) {
		_ = s.CreateProfile("staging", "staging env", false, "")
		profiles, err := s.ListProfiles()
		if err != nil {
			t.Fatalf("ListProfiles: %v", err)
		}
		if len(profiles) < 2 {
			t.Fatalf("got %d profiles, want >= 2", len(profiles))
		}
	})

	t.Run("profile exists", func(t *testing.T) {
		tests := []struct {
			name string
			want bool
		}{
			{"dev", true},
			{"staging", true},
			{"nonexistent", false},
		}
		for _, tt := range tests {
			exists, err := s.ProfileExists(tt.name)
			if err != nil {
				t.Fatalf("ProfileExists(%q): %v", tt.name, err)
			}
			if exists != tt.want {
				t.Errorf("ProfileExists(%q) = %v, want %v", tt.name, exists, tt.want)
			}
		}
	})

	t.Run("set default profile", func(t *testing.T) {
		if err := s.SetDefaultProfile("staging"); err != nil {
			t.Fatalf("SetDefaultProfile: %v", err)
		}
		p, err := s.GetDefaultProfile()
		if err != nil {
			t.Fatalf("GetDefaultProfile: %v", err)
		}
		if p.Name != "staging" {
			t.Errorf("got %q, want %q", p.Name, "staging")
		}
		old, _ := s.GetProfile("dev")
		if old.IsDefault != nil && *old.IsDefault != 0 {
			t.Errorf("old default not cleared: %v", old.IsDefault)
		}
	})

	t.Run("delete profile", func(t *testing.T) {
		if err := s.DeleteProfile("staging"); err != nil {
			t.Fatalf("DeleteProfile: %v", err)
		}
		exists, _ := s.ProfileExists("staging")
		if exists {
			t.Error("profile still exists after delete")
		}
	})
}

func TestGetNonExistentProfile(t *testing.T) {
	s := openTestStore(t)
	_, err := s.GetProfile("ghost")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestSecretCRUD(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("prod", "production", true, "")

	encVal := []byte("encrypted-data")
	nonce := []byte("nonce-12byte")

	t.Run("upsert and get", func(t *testing.T) {
		if err := s.UpsertSecret("prod", "API_KEY", encVal, nonce, "api key"); err != nil {
			t.Fatalf("UpsertSecret: %v", err)
		}
		sec, err := s.GetSecret("prod", "API_KEY")
		if err != nil {
			t.Fatalf("GetSecret: %v", err)
		}
		if sec.Key != "API_KEY" {
			t.Errorf("got key %q, want %q", sec.Key, "API_KEY")
		}
		if string(sec.EncryptedValue) != string(encVal) {
			t.Error("encrypted value mismatch")
		}
		if string(sec.Nonce) != string(nonce) {
			t.Error("nonce mismatch")
		}
	})

	t.Run("upsert overwrites", func(t *testing.T) {
		newVal := []byte("updated-data")
		if err := s.UpsertSecret("prod", "API_KEY", newVal, nonce, "updated"); err != nil {
			t.Fatalf("UpsertSecret: %v", err)
		}
		sec, _ := s.GetSecret("prod", "API_KEY")
		if string(sec.EncryptedValue) != string(newVal) {
			t.Error("upsert did not overwrite")
		}
	})

	t.Run("secret exists", func(t *testing.T) {
		tests := []struct {
			profile string
			key     string
			want    bool
		}{
			{"prod", "API_KEY", true},
			{"prod", "MISSING", false},
			{"ghost", "API_KEY", false},
		}
		for _, tt := range tests {
			exists, err := s.SecretExists(tt.profile, tt.key)
			if err != nil {
				t.Fatalf("SecretExists(%q, %q): %v", tt.profile, tt.key, err)
			}
			if exists != tt.want {
				t.Errorf("SecretExists(%q, %q) = %v, want %v", tt.profile, tt.key, exists, tt.want)
			}
		}
	})

	t.Run("count secrets", func(t *testing.T) {
		_ = s.UpsertSecret("prod", "DB_PASS", encVal, nonce, "db password")
		count, err := s.CountSecrets("prod")
		if err != nil {
			t.Fatalf("CountSecrets: %v", err)
		}
		if count != 2 {
			t.Errorf("got count %d, want 2", count)
		}
	})

	t.Run("list secrets", func(t *testing.T) {
		secrets, err := s.ListSecrets("prod")
		if err != nil {
			t.Fatalf("ListSecrets: %v", err)
		}
		if len(secrets) != 2 {
			t.Errorf("got %d secrets, want 2", len(secrets))
		}
	})

	t.Run("list all secrets", func(t *testing.T) {
		_ = s.CreateProfile("other", "other", false, "")
		_ = s.UpsertSecret("other", "TOKEN", encVal, nonce, "token")
		all, err := s.ListAllSecrets()
		if err != nil {
			t.Fatalf("ListAllSecrets: %v", err)
		}
		if len(all) != 3 {
			t.Errorf("got %d secrets, want 3", len(all))
		}
	})

	t.Run("delete secret", func(t *testing.T) {
		if err := s.DeleteSecret("prod", "API_KEY"); err != nil {
			t.Fatalf("DeleteSecret: %v", err)
		}
		exists, _ := s.SecretExists("prod", "API_KEY")
		if exists {
			t.Error("secret still exists after delete")
		}
	})
}

func TestGetNonExistentSecret(t *testing.T) {
	s := openTestStore(t)
	_, err := s.GetSecret("ghost", "nope")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteProfileWithSecrets(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("temp", "temporary", false, "")
	_ = s.UpsertSecret("temp", "KEY1", []byte("v"), []byte("n"), "")
	_ = s.UpsertSecret("temp", "KEY2", []byte("v"), []byte("n"), "")

	_ = s.DeleteSecret("temp", "KEY1")
	_ = s.DeleteSecret("temp", "KEY2")

	if err := s.DeleteProfile("temp"); err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}

	exists, _ := s.ProfileExists("temp")
	if exists {
		t.Error("profile still exists after delete")
	}
}

func TestSealedKeyCRUD(t *testing.T) {
	s := openTestStore(t)

	t.Run("no sealed key initially", func(t *testing.T) {
		has, err := s.HasSealedKey()
		if err != nil {
			t.Fatalf("HasSealedKey: %v", err)
		}
		if has {
			t.Error("expected no sealed key")
		}
	})

	sealedData := []byte("sealed-bytes")
	nonce := []byte("nonce-bytes")
	keySalt := []byte("salt-bytes")
	machineID := []byte("machine-hash")

	t.Run("upsert and get", func(t *testing.T) {
		if err := s.UpsertSealedKey(sealedData, nonce, keySalt, machineID, 1, "software"); err != nil {
			t.Fatalf("UpsertSealedKey: %v", err)
		}
		has, _ := s.HasSealedKey()
		if !has {
			t.Fatal("expected sealed key to exist")
		}
		sk, err := s.GetSealedKey()
		if err != nil {
			t.Fatalf("GetSealedKey: %v", err)
		}
		if string(sk.SealedData) != string(sealedData) {
			t.Error("sealed data mismatch")
		}
		if string(sk.Nonce) != string(nonce) {
			t.Error("nonce mismatch")
		}
		if string(sk.KeySalt) != string(keySalt) {
			t.Error("key salt mismatch")
		}
		if string(sk.MachineIDHash) != string(machineID) {
			t.Error("machine ID hash mismatch")
		}
		if sk.Version == nil || *sk.Version != 1 {
			t.Errorf("version = %v, want 1", sk.Version)
		}
		if sk.SealMethod != "software" {
			t.Errorf("seal_method = %q, want %q", sk.SealMethod, "software")
		}
	})

	t.Run("upsert overwrites", func(t *testing.T) {
		newData := []byte("new-sealed")
		if err := s.UpsertSealedKey(newData, nonce, keySalt, machineID, 2, "tpm"); err != nil {
			t.Fatalf("UpsertSealedKey: %v", err)
		}
		sk, _ := s.GetSealedKey()
		if string(sk.SealedData) != string(newData) {
			t.Error("upsert did not overwrite sealed data")
		}
		if sk.SealMethod != "tpm" {
			t.Errorf("seal_method = %q, want %q", sk.SealMethod, "tpm")
		}
	})

	t.Run("delete", func(t *testing.T) {
		if err := s.DeleteSealedKey(); err != nil {
			t.Fatalf("DeleteSealedKey: %v", err)
		}
		has, _ := s.HasSealedKey()
		if has {
			t.Error("sealed key still exists after delete")
		}
	})
}

func TestPasswordCRUD(t *testing.T) {
	s := openTestStore(t)

	t.Run("no password initially", func(t *testing.T) {
		has, err := s.HasPassword()
		if err != nil {
			t.Fatalf("HasPassword: %v", err)
		}
		if has {
			t.Error("expected no password")
		}
	})

	hash := []byte("bcrypt-hash-here")

	t.Run("upsert and get", func(t *testing.T) {
		if err := s.UpsertPassword(hash); err != nil {
			t.Fatalf("UpsertPassword: %v", err)
		}
		has, _ := s.HasPassword()
		if !has {
			t.Fatal("expected password to exist")
		}
		pw, err := s.GetPassword()
		if err != nil {
			t.Fatalf("GetPassword: %v", err)
		}
		if string(pw.PasswordHash) != string(hash) {
			t.Error("password hash mismatch")
		}
	})

	t.Run("upsert overwrites", func(t *testing.T) {
		newHash := []byte("new-bcrypt-hash")
		if err := s.UpsertPassword(newHash); err != nil {
			t.Fatalf("UpsertPassword: %v", err)
		}
		pw, _ := s.GetPassword()
		if string(pw.PasswordHash) != string(newHash) {
			t.Error("upsert did not overwrite password")
		}
	})

	t.Run("delete", func(t *testing.T) {
		if err := s.DeletePassword(); err != nil {
			t.Fatalf("DeletePassword: %v", err)
		}
		has, _ := s.HasPassword()
		if has {
			t.Error("password still exists after delete")
		}
	})
}

func TestBeginTxEndTx(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("txtest", "tx test", false, "")

	tx, q, err := s.BeginTx()
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	desc := ""
	err = q.UpsertSecret(context.Background(), sqlc.UpsertSecretParams{
		ProfileName:    "txtest",
		Key:            "TXK",
		EncryptedValue: []byte("v"),
		Nonce:          []byte("n"),
		Description:    &desc,
	})
	if err != nil {
		_ = tx.Rollback()
		s.EndTx()
		t.Fatalf("UpsertSecret in tx: %v", err)
	}

	if err := tx.Commit(); err != nil {
		s.EndTx()
		t.Fatalf("Commit: %v", err)
	}
	s.EndTx()

	exists, _ := s.SecretExists("txtest", "TXK")
	if !exists {
		t.Error("secret not found after tx commit")
	}
}

func TestBeginTxRollback(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("rollback", "rollback test", false, "")

	tx, q, err := s.BeginTx()
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	desc2 := ""
	_ = q.UpsertSecret(context.Background(), sqlc.UpsertSecretParams{
		ProfileName:    "rollback",
		Key:            "RBK",
		EncryptedValue: []byte("v"),
		Nonce:          []byte("n"),
		Description:    &desc2,
	})
	_ = tx.Rollback()
	s.EndTx()

	exists, _ := s.SecretExists("rollback", "RBK")
	if exists {
		t.Error("secret found after tx rollback")
	}
}

func TestAuditLogCRUD(t *testing.T) {
	s := openTestStore(t)

	t.Run("empty initially", func(t *testing.T) {
		count, err := s.CountAuditLog()
		if err != nil {
			t.Fatalf("CountAuditLog: %v", err)
		}
		if count != 0 {
			t.Errorf("got count %d, want 0", count)
		}
	})

	t.Run("insert and list", func(t *testing.T) {
		if err := s.LogAudit("secret.set", "dev", "API_KEY", ""); err != nil {
			t.Fatalf("LogAudit: %v", err)
		}
		if err := s.LogAudit("secret.get", "dev", "API_KEY", ""); err != nil {
			t.Fatalf("LogAudit: %v", err)
		}
		if err := s.LogAudit("profile.create", "staging", "", "new profile"); err != nil {
			t.Fatalf("LogAudit: %v", err)
		}

		entries, err := s.ListAuditLog(10)
		if err != nil {
			t.Fatalf("ListAuditLog: %v", err)
		}
		if len(entries) != 3 {
			t.Fatalf("got %d entries, want 3", len(entries))
		}
		if entries[0].Action != "profile.create" {
			t.Errorf("expected newest first, got %q", entries[0].Action)
		}
	})

	t.Run("list by profile", func(t *testing.T) {
		entries, err := s.ListAuditLogByProfile("dev", 10)
		if err != nil {
			t.Fatalf("ListAuditLogByProfile: %v", err)
		}
		if len(entries) != 2 {
			t.Errorf("got %d entries, want 2", len(entries))
		}
	})

	t.Run("list by action", func(t *testing.T) {
		entries, err := s.ListAuditLogByAction("secret.set", 10)
		if err != nil {
			t.Fatalf("ListAuditLogByAction: %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("got %d entries, want 1", len(entries))
		}
	})

	t.Run("count", func(t *testing.T) {
		count, err := s.CountAuditLog()
		if err != nil {
			t.Fatalf("CountAuditLog: %v", err)
		}
		if count != 3 {
			t.Errorf("got count %d, want 3", count)
		}
	})

	t.Run("purge", func(t *testing.T) {
		// Purge entries older than far in the future (all of them).
		future := time.Now().Add(24 * time.Hour)
		if err := s.PurgeAuditLog(future); err != nil {
			t.Fatalf("PurgeAuditLog: %v", err)
		}
		count, _ := s.CountAuditLog()
		if count != 0 {
			t.Errorf("got count %d after purge, want 0", count)
		}
	})
}

func TestSecretVersionCRUD(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("prod", "production", true, "")
	_ = s.UpsertSecret("prod", "API_KEY", []byte("v1-enc"), []byte("v1-nonce"), "api key")

	t.Run("insert and list versions", func(t *testing.T) {
		if err := s.InsertSecretVersion("prod", "API_KEY", 1, []byte("v1-enc"), []byte("v1-nonce"), time.Now().Add(30*24*time.Hour)); err != nil {
			t.Fatalf("InsertSecretVersion: %v", err)
		}
		if err := s.InsertSecretVersion("prod", "API_KEY", 2, []byte("v2-enc"), []byte("v2-nonce"), time.Now().Add(30*24*time.Hour)); err != nil {
			t.Fatalf("InsertSecretVersion: %v", err)
		}

		versions, err := s.ListSecretVersions("prod", "API_KEY")
		if err != nil {
			t.Fatalf("ListSecretVersions: %v", err)
		}
		if len(versions) != 2 {
			t.Fatalf("got %d versions, want 2", len(versions))
		}
		if versions[0].Version != 2 {
			t.Errorf("first version = %d, want 2 (newest first)", versions[0].Version)
		}
	})

	t.Run("get specific version", func(t *testing.T) {
		ver, err := s.GetSecretVersion("prod", "API_KEY", 1)
		if err != nil {
			t.Fatalf("GetSecretVersion: %v", err)
		}
		if string(ver.EncryptedValue) != "v1-enc" {
			t.Errorf("encrypted value = %q, want %q", ver.EncryptedValue, "v1-enc")
		}
	})

	t.Run("get latest version number", func(t *testing.T) {
		num, err := s.GetLatestVersionNumber("prod", "API_KEY")
		if err != nil {
			t.Fatalf("GetLatestVersionNumber: %v", err)
		}
		if num != 2 {
			t.Errorf("latest version = %d, want 2", num)
		}
	})

	t.Run("get latest version number for nonexistent key", func(t *testing.T) {
		num, err := s.GetLatestVersionNumber("prod", "NOPE")
		if err != nil {
			t.Fatalf("GetLatestVersionNumber: %v", err)
		}
		if num != 0 {
			t.Errorf("latest version = %d, want 0", num)
		}
	})

	t.Run("delete versions", func(t *testing.T) {
		if err := s.DeleteSecretVersions("prod", "API_KEY"); err != nil {
			t.Fatalf("DeleteSecretVersions: %v", err)
		}
		versions, err := s.ListSecretVersions("prod", "API_KEY")
		if err != nil {
			t.Fatalf("ListSecretVersions: %v", err)
		}
		if len(versions) != 0 {
			t.Errorf("got %d versions after delete, want 0", len(versions))
		}
	})

	t.Run("list all secret versions", func(t *testing.T) {
		_ = s.CreateProfile("stg", "staging", false, "")
		_ = s.UpsertSecret("stg", "TOKEN", []byte("enc"), []byte("nonce"), "")
		_ = s.InsertSecretVersion("prod", "API_KEY", 1, []byte("enc1"), []byte("n1"), time.Now().Add(30*24*time.Hour))
		_ = s.InsertSecretVersion("stg", "TOKEN", 1, []byte("enc2"), []byte("n2"), time.Now().Add(30*24*time.Hour))

		all, err := s.ListAllSecretVersions()
		if err != nil {
			t.Fatalf("ListAllSecretVersions: %v", err)
		}
		if len(all) != 2 {
			t.Errorf("got %d versions, want 2", len(all))
		}
	})
}

func TestProfileUUIDMethods(t *testing.T) {
	s := openTestStore(t)

	t.Run("create with UUID and get by UUID", func(t *testing.T) {
		if err := s.CreateProfile("uuid-test", "test", false, "test-uuid-1234"); err != nil {
			t.Fatalf("CreateProfile: %v", err)
		}

		p, err := s.GetProfileByUUID("test-uuid-1234")
		if err != nil {
			t.Fatalf("GetProfileByUUID: %v", err)
		}
		if p.Name != "uuid-test" {
			t.Errorf("got name %q, want %q", p.Name, "uuid-test")
		}
	})

	t.Run("profile exists by UUID", func(t *testing.T) {
		exists, err := s.ProfileExistsByUUID("test-uuid-1234")
		if err != nil {
			t.Fatalf("ProfileExistsByUUID: %v", err)
		}
		if !exists {
			t.Error("expected profile to exist")
		}

		exists, err = s.ProfileExistsByUUID("nonexistent")
		if err != nil {
			t.Fatalf("ProfileExistsByUUID: %v", err)
		}
		if exists {
			t.Error("expected profile not to exist")
		}
	})

	t.Run("get profile UUID by name", func(t *testing.T) {
		uuid, err := s.GetProfileUUID("uuid-test")
		if err != nil {
			t.Fatalf("GetProfileUUID: %v", err)
		}
		if uuid != "test-uuid-1234" {
			t.Errorf("got UUID %q, want %q", uuid, "test-uuid-1234")
		}
	})

	t.Run("get UUID for nonexistent profile", func(t *testing.T) {
		_, err := s.GetProfileUUID("ghost")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("get profile by nonexistent UUID", func(t *testing.T) {
		_, err := s.GetProfileByUUID("no-such-uuid")
		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestTemplateCRUD(t *testing.T) {
	s := openTestStore(t)

	t.Run("create and get", func(t *testing.T) {
		if err := s.CreateTemplate("db-url", "Database URL template", `postgres://{{.user}}:{{.pass}}@{{.host}}/{{.db}}`, false); err != nil {
			t.Fatalf("CreateTemplate: %v", err)
		}

		tmpl, err := s.GetTemplate("db-url")
		if err != nil {
			t.Fatalf("GetTemplate: %v", err)
		}
		if tmpl.Name != "db-url" {
			t.Errorf("got name %q, want %q", tmpl.Name, "db-url")
		}
	})

	t.Run("list templates", func(t *testing.T) {
		_ = s.CreateTemplate("api-key", "API key template", `{{.key}}`, true)

		templates, err := s.ListTemplates()
		if err != nil {
			t.Fatalf("ListTemplates: %v", err)
		}
		if len(templates) < 2 {
			t.Fatalf("got %d templates, want >= 2", len(templates))
		}
	})

	t.Run("template exists", func(t *testing.T) {
		exists, err := s.TemplateExists("db-url")
		if err != nil {
			t.Fatalf("TemplateExists: %v", err)
		}
		if !exists {
			t.Error("expected template to exist")
		}

		exists, err = s.TemplateExists("nonexistent")
		if err != nil {
			t.Fatalf("TemplateExists: %v", err)
		}
		if exists {
			t.Error("expected template not to exist")
		}
	})

	t.Run("delete template", func(t *testing.T) {
		if err := s.DeleteTemplate("db-url"); err != nil {
			t.Fatalf("DeleteTemplate: %v", err)
		}
		exists, _ := s.TemplateExists("db-url")
		if exists {
			t.Error("template still exists after delete")
		}
	})
}

func TestPurgeExpiredVersions(t *testing.T) {
	s := openTestStore(t)
	_ = s.CreateProfile("purge-test", "purge test", true, "")
	_ = s.UpsertSecret("purge-test", "KEY1", []byte("enc"), []byte("nonce"), "key")

	// Insert an already-expired version.
	expired := time.Now().Add(-1 * time.Hour)
	if err := s.InsertSecretVersion("purge-test", "KEY1", 1, []byte("old-enc"), []byte("old-n"), expired); err != nil {
		t.Fatalf("InsertSecretVersion (expired): %v", err)
	}

	// Insert a still-valid version.
	future := time.Now().Add(30 * 24 * time.Hour)
	if err := s.InsertSecretVersion("purge-test", "KEY1", 2, []byte("new-enc"), []byte("new-n"), future); err != nil {
		t.Fatalf("InsertSecretVersion (valid): %v", err)
	}

	count, err := s.PurgeExpiredVersions()
	if err != nil {
		t.Fatalf("PurgeExpiredVersions: %v", err)
	}
	if count != 1 {
		t.Errorf("purged %d versions, want 1", count)
	}

	versions, err := s.ListSecretVersions("purge-test", "KEY1")
	if err != nil {
		t.Fatalf("ListSecretVersions: %v", err)
	}
	if len(versions) != 1 {
		t.Errorf("got %d versions after purge, want 1", len(versions))
	}
	if versions[0].Version != 2 {
		t.Errorf("remaining version = %d, want 2", versions[0].Version)
	}
}

func TestClosePreventsFurtherOps(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := s.Ping(); err == nil {
		t.Error("expected error after Close, got nil")
	}
}
