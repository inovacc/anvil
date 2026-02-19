package vault_test

import (
	"testing"
)

func TestRegisterAndListApps(t *testing.T) {
	v, _ := initAndOpen(t)

	info, err := v.RegisterApp("myapp", "my application", "svc-001")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	if info.Name != "myapp" {
		t.Errorf("name = %q, want %q", info.Name, "myapp")
	}

	if info.UUID == "" {
		t.Error("expected non-empty UUID")
	}

	if info.Status != "active" {
		t.Errorf("status = %q, want %q", info.Status, "active")
	}

	apps, err := v.ListApps()
	if err != nil {
		t.Fatalf("ListApps: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("len = %d, want 1", len(apps))
	}

	if apps[0].Name != "myapp" {
		t.Errorf("name = %q, want %q", apps[0].Name, "myapp")
	}
}

func TestRegisterDuplicateApp(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("dup", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	_, err = v.RegisterApp("dup", "", "")
	if err == nil {
		t.Fatal("expected error for duplicate app name")
	}
}

func TestGetAppByNameAndUUID(t *testing.T) {
	v, _ := initAndOpen(t)

	info, err := v.RegisterApp("lookup", "test", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	byName, err := v.GetApp("lookup")
	if err != nil {
		t.Fatalf("GetApp by name: %v", err)
	}

	if byName.UUID != info.UUID {
		t.Errorf("UUID mismatch: %q != %q", byName.UUID, info.UUID)
	}

	byUUID, err := v.GetApp(info.UUID)
	if err != nil {
		t.Fatalf("GetApp by UUID: %v", err)
	}

	if byUUID.Name != "lookup" {
		t.Errorf("name = %q, want %q", byUUID.Name, "lookup")
	}
}

func TestGetAppNotFound(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.GetApp("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent app")
	}
}

func TestOpenAppSetGetDelete(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("secrets-test", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	av, err := v.OpenApp("secrets-test")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}

	defer func() { _ = av.Close() }()

	if err := av.Set("DB_HOST", "localhost", "database host"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	val, err := av.Get("DB_HOST")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if val != "localhost" {
		t.Errorf("value = %q, want %q", val, "localhost")
	}

	secrets, err := av.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(secrets) != 1 {
		t.Errorf("len = %d, want 1", len(secrets))
	}

	if err := av.Delete("DB_HOST"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = av.Get("DB_HOST")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestOpenAppExportImport(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("export-test", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	av, err := v.OpenApp("export-test")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}

	if err := av.Set("KEY1", "val1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := av.Set("KEY2", "val2", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	entries, err := av.Export()
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("len = %d, want 2", len(entries))
	}

	_ = av.Close()

	// Import into another app.
	_, err = v.RegisterApp("import-test", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	av2, err := v.OpenApp("import-test")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}

	defer func() { _ = av2.Close() }()

	if err := av2.Import(entries); err != nil {
		t.Fatalf("Import: %v", err)
	}

	imported, err := av2.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(imported) != 2 {
		t.Errorf("len = %d, want 2", len(imported))
	}
}

func TestDisableEnableApp(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("toggle", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	if err := v.DisableApp("toggle"); err != nil {
		t.Fatalf("DisableApp: %v", err)
	}

	info, err := v.GetApp("toggle")
	if err != nil {
		t.Fatalf("GetApp: %v", err)
	}

	if info.Status != "disabled" {
		t.Errorf("status = %q, want %q", info.Status, "disabled")
	}

	// Opening disabled app should fail.
	_, err = v.OpenApp("toggle")
	if err == nil {
		t.Fatal("expected error opening disabled app")
	}

	if err := v.EnableApp("toggle"); err != nil {
		t.Fatalf("EnableApp: %v", err)
	}

	info, err = v.GetApp("toggle")
	if err != nil {
		t.Fatalf("GetApp: %v", err)
	}

	if info.Status != "active" {
		t.Errorf("status = %q, want %q", info.Status, "active")
	}
}

func TestRemoveApp(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("removeme", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	if err := v.RemoveApp("removeme"); err != nil {
		t.Fatalf("RemoveApp: %v", err)
	}

	apps, err := v.ListApps()
	if err != nil {
		t.Fatalf("ListApps: %v", err)
	}

	if len(apps) != 0 {
		t.Errorf("len = %d, want 0", len(apps))
	}
}

func TestAppSecretCountUpdate(t *testing.T) {
	v, _ := initAndOpen(t)

	_, err := v.RegisterApp("count-test", "", "")
	if err != nil {
		t.Fatalf("RegisterApp: %v", err)
	}

	av, err := v.OpenApp("count-test")
	if err != nil {
		t.Fatalf("OpenApp: %v", err)
	}

	if err := av.Set("K1", "v1", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	if err := av.Set("K2", "v2", ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_ = av.Close()

	info, err := v.GetApp("count-test")
	if err != nil {
		t.Fatalf("GetApp: %v", err)
	}

	if info.SecretCount != 2 {
		t.Errorf("secret_count = %d, want 2", info.SecretCount)
	}
}
