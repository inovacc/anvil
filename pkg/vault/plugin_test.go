package vault_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
)

func writeTestScript(t *testing.T, dir, name, content string) string {
	t.Helper()

	var scriptPath string
	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(dir, name+".bat")
		if err := os.WriteFile(scriptPath, []byte(content), 0o700); err != nil {
			t.Fatalf("write script: %v", err)
		}
	} else {
		scriptPath = filepath.Join(dir, name+".sh")
		if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\n"+content), 0o700); err != nil {
			t.Fatalf("write script: %v", err)
		}
	}

	return scriptPath
}

func TestPluginManagerNoConfig(t *testing.T) {
	pm := vault.NewPluginManager("/nonexistent/path/plugins.json")

	// Should be a no-op manager.
	if pm.HasHooks(vault.HookPreSet) {
		t.Error("expected no hooks")
	}

	providers := pm.ListProviders()
	if len(providers) != 0 {
		t.Errorf("expected no providers, got %d", len(providers))
	}

	// RunHooks should be no-op.
	if err := pm.RunHooks(vault.HookPreSet, vault.HookPayload{}); err != nil {
		t.Errorf("RunHooks: %v", err)
	}
}

func TestPluginManagerLoadConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "plugins.json")

	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "pre-set", "command": "echo", "args": []string{"hello"}},
		},
		"providers": []map[string]any{
			{"name": "test-provider", "command": "echo", "prefix": "ext/"},
		},
	}

	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	pm := vault.NewPluginManager(configPath)

	if !pm.HasHooks(vault.HookPreSet) {
		t.Error("expected hook for pre-set")
	}

	if pm.HasHooks(vault.HookPostSet) {
		t.Error("expected no hook for post-set")
	}

	providers := pm.ListProviders()
	if len(providers) != 1 || providers[0] != "test-provider" {
		t.Errorf("expected [test-provider], got %v", providers)
	}
}

func TestPluginManagerInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "plugins.json")

	if err := os.WriteFile(configPath, []byte("not json"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	pm := vault.NewPluginManager(configPath)

	// Should fall back to empty config.
	if pm.HasHooks(vault.HookPreSet) {
		t.Error("expected no hooks from invalid config")
	}
}

func TestPluginManagerAddRemoveHook(t *testing.T) {
	pm := vault.NewPluginManager("/nonexistent")

	pm.AddHook(vault.HookPreSet, "echo", []string{"test"})
	if !pm.HasHooks(vault.HookPreSet) {
		t.Error("expected hook after add")
	}

	pm.RemoveHook(vault.HookPreSet, "echo")
	if pm.HasHooks(vault.HookPreSet) {
		t.Error("expected no hook after remove")
	}
}

func TestPluginManagerAddRemoveProvider(t *testing.T) {
	pm := vault.NewPluginManager("/nonexistent")

	pm.AddProvider("aws", "aws-provider", "aws/")
	providers := pm.ListProviders()
	if len(providers) != 1 || providers[0] != "aws" {
		t.Errorf("expected [aws], got %v", providers)
	}

	pm.RemoveProvider("aws")
	providers = pm.ListProviders()
	if len(providers) != 0 {
		t.Errorf("expected empty, got %v", providers)
	}
}

func TestPluginManagerSaveConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "subdir", "plugins.json")

	pm := vault.NewPluginManager(configPath)
	pm.AddHook(vault.HookPostSet, "notify", nil)
	pm.AddProvider("vault", "hcvault", "hcv/")

	if err := pm.SaveConfig(); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Reload and verify.
	pm2 := vault.NewPluginManager(configPath)
	if !pm2.HasHooks(vault.HookPostSet) {
		t.Error("expected hook after reload")
	}

	if providers := pm2.ListProviders(); len(providers) != 1 {
		t.Errorf("expected 1 provider after reload, got %d", len(providers))
	}
}

func TestPluginManagerConfig(t *testing.T) {
	pm := vault.NewPluginManager("/nonexistent")
	cfg := pm.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}

	// Nil manager should return empty config.
	var nilPM *vault.PluginManager
	cfg = nilPM.Config()
	if cfg == nil {
		t.Fatal("nil manager Config() returned nil")
	}
}

func TestPluginManagerRunHooksPostEvent(t *testing.T) {
	dir := t.TempDir()

	// Create a script that outputs valid JSON with allow=true.
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = writeTestScript(t, dir, "hook", `@echo off
echo {"allow":true}`)
	} else {
		cmd = writeTestScript(t, dir, "hook", `echo '{"allow":true}'`)
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "post-set", "command": cmd},
		},
	}

	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)

	err := pm.RunHooks(vault.HookPostSet, vault.HookPayload{
		SecretKey:   "MY_KEY",
		ProfileName: "default",
	})
	if err != nil {
		t.Errorf("RunHooks post-set: %v", err)
	}
}

func TestPluginManagerRunHooksPreBlock(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = writeTestScript(t, dir, "hook", `@echo off
echo {"allow":false,"message":"blocked by policy"}`)
	} else {
		cmd = writeTestScript(t, dir, "hook", `echo '{"allow":false,"message":"blocked by policy"}'`)
	}

	configPath := filepath.Join(dir, "plugins.json")
	cfg := map[string]any{
		"hooks": []map[string]any{
			{"event": "pre-set", "command": cmd},
		},
	}

	data, _ := json.Marshal(cfg)
	_ = os.WriteFile(configPath, data, 0o600)

	pm := vault.NewPluginManager(configPath)

	err := pm.RunHooks(vault.HookPreSet, vault.HookPayload{
		SecretKey:   "BLOCKED_KEY",
		ProfileName: "default",
	})
	if err == nil {
		t.Fatal("expected error from blocking pre-hook")
	}

	if msg := err.Error(); msg != "blocked by policy" {
		t.Errorf("expected 'blocked by policy', got %q", msg)
	}
}

func TestPluginManagerGetFromProviderNotFound(t *testing.T) {
	pm := vault.NewPluginManager("/nonexistent")

	_, err := pm.GetFromProvider("missing", "key", "profile")
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestPluginManagerNilRunHooks(t *testing.T) {
	var pm *vault.PluginManager

	if err := pm.RunHooks(vault.HookPreSet, vault.HookPayload{}); err != nil {
		t.Errorf("nil RunHooks: %v", err)
	}
}
