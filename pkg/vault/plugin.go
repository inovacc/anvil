package vault

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// HookEvent represents a vault lifecycle event that can trigger hooks.
type HookEvent string

const (
	HookPreSet     HookEvent = "pre-set"
	HookPostSet    HookEvent = "post-set"
	HookPreGet     HookEvent = "pre-get"
	HookPostGet    HookEvent = "post-get"
	HookPreDelete  HookEvent = "pre-delete"
	HookPostDelete HookEvent = "post-delete"
	HookPostRotate HookEvent = "post-rotate"
	HookPostInit   HookEvent = "post-init"
)

// HookPayload is the JSON data sent to hook executables via stdin.
type HookPayload struct {
	Event       HookEvent `json:"event"`
	ProfileName string    `json:"profile_name,omitempty"`
	SecretKey   string    `json:"secret_key,omitempty"`
	DBPath      string    `json:"db_path,omitempty"`
}

// HookResult is the JSON response from a hook executable.
type HookResult struct {
	Allow   bool   `json:"allow"`
	Message string `json:"message,omitempty"`
}

// PluginConfig represents the plugin configuration for a vault.
type PluginConfig struct {
	Hooks     []HookConfig     `json:"hooks,omitempty" yaml:"hooks,omitempty"`
	Providers []ProviderConfig `json:"providers,omitempty" yaml:"providers,omitempty"`
}

// HookConfig defines a hook binding: an event mapped to an executable.
type HookConfig struct {
	Event   HookEvent `json:"event" yaml:"event"`
	Command string    `json:"command" yaml:"command"`
	Args    []string  `json:"args,omitempty" yaml:"args,omitempty"`
}

// ProviderConfig defines an external secret provider.
type ProviderConfig struct {
	Name    string `json:"name" yaml:"name"`
	Command string `json:"command" yaml:"command"`
	Prefix  string `json:"prefix" yaml:"prefix"`
}

// ProviderRequest is sent to a secret provider via stdin.
type ProviderRequest struct {
	Action      string `json:"action"`
	Key         string `json:"key,omitempty"`
	ProfileName string `json:"profile_name,omitempty"`
}

// ProviderResponse is the JSON response from a secret provider.
type ProviderResponse struct {
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

// PluginManager manages hooks and providers for a vault.
type PluginManager struct {
	config     *PluginConfig
	configPath string
}

// NewPluginManager creates a plugin manager from a config file path.
// If the config file doesn't exist, an empty manager is returned (no-op).
func NewPluginManager(configPath string) *PluginManager {
	pm := &PluginManager{
		config:     &PluginConfig{},
		configPath: configPath,
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return pm // No config file, no plugins.
	}

	var cfg PluginConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		slog.Error("failed to parse plugin config", "path", configPath, "error", err)
		return pm
	}

	pm.config = &cfg

	return pm
}

// RunHooks executes all hooks registered for the given event.
// Pre-hooks can block the operation by returning Allow=false.
// Errors are logged but never block operations (best-effort).
func (pm *PluginManager) RunHooks(event HookEvent, payload HookPayload) error {
	if pm == nil || pm.config == nil {
		return nil
	}

	payload.Event = event

	for _, hook := range pm.config.Hooks {
		if hook.Event != event {
			continue
		}

		result, err := executeHook(hook, payload)
		if err != nil {
			slog.Error("hook execution failed", "event", event, "command", hook.Command, "error", err)
			continue
		}

		// Pre-hooks can block operations.
		if strings.HasPrefix(string(event), "pre-") && !result.Allow {
			msg := result.Message
			if msg == "" {
				msg = fmt.Sprintf("blocked by hook %q", hook.Command)
			}

			return &UserError{
				Message: msg,
				Hint:    "Check your plugin configuration or remove the hook",
			}
		}
	}

	return nil
}

// GetFromProvider queries an external provider for a secret value.
func (pm *PluginManager) GetFromProvider(providerName, key, profileName string) (string, error) {
	if pm == nil || pm.config == nil {
		return "", fmt.Errorf("no providers configured")
	}

	for _, p := range pm.config.Providers {
		if p.Name != providerName {
			continue
		}

		return executeProvider(p, ProviderRequest{
			Action:      "get",
			Key:         key,
			ProfileName: profileName,
		})
	}

	return "", fmt.Errorf("provider %q not found", providerName)
}

// ListProviders returns the names of all configured providers.
func (pm *PluginManager) ListProviders() []string {
	if pm == nil || pm.config == nil {
		return nil
	}

	names := make([]string, 0, len(pm.config.Providers))
	for _, p := range pm.config.Providers {
		names = append(names, p.Name)
	}

	return names
}

// HasHooks returns true if any hooks are configured for the given event.
func (pm *PluginManager) HasHooks(event HookEvent) bool {
	if pm == nil || pm.config == nil {
		return false
	}

	for _, h := range pm.config.Hooks {
		if h.Event == event {
			return true
		}
	}

	return false
}

// SaveConfig writes the current plugin configuration to disk.
func (pm *PluginManager) SaveConfig() error {
	data, err := json.MarshalIndent(pm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(pm.configPath), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	return os.WriteFile(pm.configPath, data, 0o600)
}

// AddHook registers a new hook in the config.
func (pm *PluginManager) AddHook(event HookEvent, command string, args []string) {
	pm.config.Hooks = append(pm.config.Hooks, HookConfig{
		Event:   event,
		Command: command,
		Args:    args,
	})
}

// AddProvider registers a new provider in the config.
func (pm *PluginManager) AddProvider(name, command, prefix string) {
	pm.config.Providers = append(pm.config.Providers, ProviderConfig{
		Name:    name,
		Command: command,
		Prefix:  prefix,
	})
}

// RemoveHook removes all hooks for the given event and command.
func (pm *PluginManager) RemoveHook(event HookEvent, command string) {
	filtered := make([]HookConfig, 0, len(pm.config.Hooks))
	for _, h := range pm.config.Hooks {
		if h.Event == event && h.Command == command {
			continue
		}

		filtered = append(filtered, h)
	}

	pm.config.Hooks = filtered
}

// RemoveProvider removes a provider by name.
func (pm *PluginManager) RemoveProvider(name string) {
	filtered := make([]ProviderConfig, 0, len(pm.config.Providers))
	for _, p := range pm.config.Providers {
		if p.Name == name {
			continue
		}

		filtered = append(filtered, p)
	}

	pm.config.Providers = filtered
}

// Config returns the current plugin configuration (read-only).
func (pm *PluginManager) Config() *PluginConfig {
	if pm == nil {
		return &PluginConfig{}
	}

	return pm.config
}

func executeHook(hook HookConfig, payload HookPayload) (*HookResult, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	cmd := exec.Command(hook.Command, hook.Args...) //nolint:gosec
	cmd.Stdin = bytes.NewReader(data)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run hook: %w (stderr: %s)", err, stderr.String())
	}

	// If stdout is empty, assume the hook allows the operation.
	if stdout.Len() == 0 {
		return &HookResult{Allow: true}, nil
	}

	var result HookResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return &HookResult{Allow: true, Message: stdout.String()}, nil //nolint:nilerr // non-JSON treated as allow
	}

	return &result, nil
}

func executeProvider(provider ProviderConfig, request ProviderRequest) (string, error) {
	data, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	cmd := exec.Command(provider.Command) //nolint:gosec
	cmd.Stdin = bytes.NewReader(data)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run provider %q: %w (stderr: %s)", provider.Name, err, stderr.String())
	}

	var resp ProviderResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return strings.TrimSpace(stdout.String()), nil //nolint:nilerr // raw stdout as value
	}

	if resp.Error != "" {
		return "", fmt.Errorf("provider %q: %s", provider.Name, resp.Error)
	}

	return resp.Value, nil
}
