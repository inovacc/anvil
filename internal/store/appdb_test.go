package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
)

func openTestAppStore(t *testing.T) *AppStore {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "app_test.db")

	s, err := OpenApp(dbPath)
	if err != nil {
		t.Fatalf("OpenApp(%q): %v", dbPath, err)
	}

	t.Cleanup(func() { _ = s.Close() })

	return s
}

func TestAppStoreSetGet(t *testing.T) {
	s := openTestAppStore(t)

	if err := s.Set("DB_HOST", []byte("encrypted"), []byte("nonce"), "database host"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	sec, err := s.Get("DB_HOST")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if sec.Key != "DB_HOST" {
		t.Errorf("key = %q, want %q", sec.Key, "DB_HOST")
	}

	if string(sec.EncryptedValue) != "encrypted" {
		t.Errorf("value = %q, want %q", sec.EncryptedValue, "encrypted")
	}

	if sec.Description != "database host" {
		t.Errorf("description = %q, want %q", sec.Description, "database host")
	}
}

func TestAppStoreUpsert(t *testing.T) {
	s := openTestAppStore(t)

	if err := s.Set("KEY", []byte("v1"), []byte("n1"), ""); err != nil {
		t.Fatalf("Set v1: %v", err)
	}

	if err := s.Set("KEY", []byte("v2"), []byte("n2"), "updated"); err != nil {
		t.Fatalf("Set v2: %v", err)
	}

	sec, err := s.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if string(sec.EncryptedValue) != "v2" {
		t.Errorf("value = %q, want %q", sec.EncryptedValue, "v2")
	}
}

func TestAppStoreDelete(t *testing.T) {
	s := openTestAppStore(t)

	if err := s.Set("KEY", []byte("val"), []byte("n"), ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := s.Delete("KEY"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.Get("KEY")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestAppStoreDeleteNotFound(t *testing.T) {
	s := openTestAppStore(t)

	err := s.Delete("NONEXISTENT")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}

func TestAppStoreList(t *testing.T) {
	s := openTestAppStore(t)

	for _, key := range []string{"A_KEY", "B_KEY", "C_KEY"} {
		if err := s.Set(key, []byte("val"), []byte("n"), ""); err != nil {
			t.Fatalf("Set %s: %v", key, err)
		}
	}

	secrets, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(secrets) != 3 {
		t.Errorf("len = %d, want 3", len(secrets))
	}
}

func TestAppStoreCount(t *testing.T) {
	s := openTestAppStore(t)

	count, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	if err := s.Set("KEY", []byte("val"), []byte("n"), ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	count, err = s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}

	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
}

func TestAppStoreUpdateSecret(t *testing.T) {
	s := openTestAppStore(t)

	if err := s.Set("KEY", []byte("old"), []byte("n1"), ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	sec, err := s.Get("KEY")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if err := s.UpdateSecret(sec.ID, []byte("new"), []byte("n2")); err != nil {
		t.Fatalf("UpdateSecret: %v", err)
	}

	sec, err = s.Get("KEY")
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}

	if string(sec.EncryptedValue) != "new" {
		t.Errorf("value = %q, want %q", sec.EncryptedValue, "new")
	}
}
