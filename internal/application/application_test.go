package application

import (
	"runtime"
	"testing"
)

func TestGetApplicationDirectory(t *testing.T) {
	dir, err := GetApplicationDirectory()
	if err != nil {
		t.Fatalf("GetApplicationDirectory: %v", err)
	}

	if dir == "" {
		t.Fatal("expected non-empty directory")
	}

	// Must end with the app name.
	if len(dir) < len(AppName) {
		t.Fatalf("directory too short: %q", dir)
	}

	suffix := dir[len(dir)-len(AppName):]
	if suffix != AppName {
		t.Errorf("expected dir to end with %q, got %q", AppName, dir)
	}
}

func TestGetApplicationDirectoryIdempotent(t *testing.T) {
	dir1, err1 := GetApplicationDirectory()
	dir2, err2 := GetApplicationDirectory()

	if err1 != nil || err2 != nil {
		t.Fatalf("errors: %v, %v", err1, err2)
	}

	if dir1 != dir2 {
		t.Errorf("not idempotent: %q != %q", dir1, dir2)
	}
}

func TestConstants(t *testing.T) {
	if AppName != "anvil" {
		t.Errorf("AppName = %q, want %q", AppName, "anvil")
	}

	if AppExeName != "anvil" {
		t.Errorf("AppExeName = %q, want %q", AppExeName, "anvil")
	}

	if AppExeNameWindows != "anvil.exe" {
		t.Errorf("AppExeNameWindows = %q, want %q", AppExeNameWindows, "anvil.exe")
	}
}

func TestDirectoryContainsPlatformPath(t *testing.T) {
	dir, err := GetApplicationDirectory()
	if err != nil {
		t.Fatalf("GetApplicationDirectory: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		// On Windows, should use AppData\Local via UserCacheDir.
		if len(dir) < 10 {
			t.Errorf("unexpectedly short Windows path: %q", dir)
		}
	default:
		// On Linux/macOS, should use .config via UserConfigDir.
		if len(dir) < 5 {
			t.Errorf("unexpectedly short path: %q", dir)
		}
	}
}
