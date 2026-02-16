package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

// --- aicontext tests ---

func TestAicontextCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "aicontext")
	if stderr != "" {
		t.Fatalf("aicontext error: %s", stderr)
	}
	if !strings.Contains(stdout, "CLI") {
		t.Errorf("expected CLI markdown output, got: %q", stdout)
	}
}

func TestAicontextCompactCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "aicontext", "--compact")
	if stderr != "" {
		t.Fatalf("aicontext compact error: %s", stderr)
	}
	if !strings.Contains(stdout, "CLI") {
		t.Errorf("expected compact output, got: %q", stdout)
	}
	// Compact should not have "## Commands" section
	if strings.Contains(stdout, "## Commands") {
		t.Error("compact mode should not include ## Commands header")
	}
}

func TestAicontextCategoryCLI(t *testing.T) {
	setupTestVault(t)

	// Run without category to verify commands are collected
	stdout, stderr := execCmd(t, "aicontext")
	if stderr != "" {
		t.Fatalf("aicontext error: %s", stderr)
	}
	if !strings.Contains(stdout, "vault") {
		t.Errorf("expected vault in output, got: %q", stdout[:min(len(stdout), 200)])
	}
}

func TestAicontextJSONCLI(t *testing.T) {
	setupTestVault(t)

	// Without category filter, should get all commands
	stdout, _ := execCmd(t, "aicontext", "--json")
	var commands []aiCommandInfo
	if err := json.Unmarshal([]byte(stdout), &commands); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if len(commands) == 0 {
		t.Error("expected at least 1 command")
	}
}

func TestAicontextCategoryFilterEmpty(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "aicontext", "--category", "nonexistent_category_xyz", "--json")
	var commands []aiCommandInfo
	if err := json.Unmarshal([]byte(stdout), &commands); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if len(commands) != 0 {
		t.Errorf("expected 0 commands for nonexistent category, got %d", len(commands))
	}
}

// --- env-inline tests ---

func TestEnvInlineCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "INLINE_KEY", "inline_val", "-p", "test")
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")
	execCmd(t, "vault", "env", "release", "-p", "test", "--password", "mypassword1", "--ttl", "5m")
	defer execCmd(t, "vault", "env", "revoke")

	stdout, stderr := execCmd(t, "--env-inline", "INLINE_KEY", "-p", "test")
	if stderr != "" {
		t.Fatalf("env-inline error: %s", stderr)
	}
	if !strings.Contains(stdout, "inline_val") {
		t.Errorf("expected inline_val, got: %q", stdout)
	}
}

func TestEnvInlineJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "INLINE_J", "jval", "-p", "test")
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")
	execCmd(t, "vault", "env", "release", "-p", "test", "--password", "mypassword1", "--ttl", "5m")
	defer execCmd(t, "vault", "env", "revoke")

	stdout, stderr := execCmd(t, "--env-inline", "INLINE_J", "-p", "test", "--json")
	if stderr != "" {
		t.Fatalf("env-inline JSON error: %s", stderr)
	}
	var result struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}
	if result.Value != "jval" {
		t.Errorf("value = %q, want %q", result.Value, "jval")
	}
}

// --- password validation tests ---

func TestPasswordTooShortCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "env", "password", "set", "--password", "short")
	if !strings.Contains(stderr, "8 characters") {
		t.Errorf("expected min length error, got: %q", stderr)
	}
}

func TestPasswordUpdateWrongCurrentCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	_, stderr := execCmd(t, "vault", "env", "password", "set", "--password", "newpassword1", "--current", "wrongpass1")
	if stderr == "" {
		t.Error("expected error for wrong current password")
	}
}

func TestPasswordUpdateMissingCurrentCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	_, stderr := execCmd(t, "vault", "env", "password", "set", "--password", "newpassword1")
	// Without --current, the command should fail (either --current required or incorrect password)
	if stderr == "" {
		t.Error("expected error when updating without --current")
	}
}

// --- completion tests ---

func TestCompleteProfileNames(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "comp-test", "--default")

	names, directive := completeProfileNames(nil, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("expected NoFileComp directive")
	}
	found := false
	for _, n := range names {
		if n == "comp-test" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected comp-test in completions, got: %v", names)
	}
}

func TestCompleteProfileNamesAlreadyHasArgs(t *testing.T) {
	names, directive := completeProfileNames(nil, []string{"existing"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("expected NoFileComp directive")
	}
	if len(names) != 0 {
		t.Errorf("expected empty completions when args present, got: %v", names)
	}
}

func TestCompleteSecretKeys(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "comp-secrets", "--default")
	execCmd(t, "vault", "set", "COMP_KEY_1", "v1", "-p", "comp-secrets")

	cmd := &cobra.Command{}
	cmd.Flags().StringP("profile", "p", "comp-secrets", "")

	keys, directive := completeSecretKeys(cmd, nil, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("expected NoFileComp directive")
	}
	found := false
	for _, k := range keys {
		if k == "COMP_KEY_1" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected COMP_KEY_1 in completions, got: %v", keys)
	}
}

func TestCompleteSecretKeysAlreadyHasArgs(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringP("profile", "p", "", "")

	keys, directive := completeSecretKeys(cmd, []string{"existing"}, "")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("expected NoFileComp directive")
	}
	if len(keys) != 0 {
		t.Errorf("expected empty completions when args present, got: %v", keys)
	}
}

// --- output/error helper tests ---

func TestOutputResultText(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	outputResult(cmd, struct{ X int }{42}, func() {
		_, _ = out.WriteString("hello text")
	})

	if out.String() != "hello text" {
		t.Errorf("text output = %q, want %q", out.String(), "hello text")
	}
}

func TestOutputResultJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	_ = cmd.Flags().Set("json", "true")
	out := &bytes.Buffer{}
	cmd.SetOut(out)

	data := struct {
		Name string `json:"name"`
	}{Name: "test"}
	outputResult(cmd, data, func() {
		t.Error("text fn should not be called in JSON mode")
	})

	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
}

func TestHandleErrorUserError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)

	ue := &vault.UserError{Message: "bad input", Hint: "try again"}
	handleError(cmd, ue)

	output := errBuf.String()
	if !strings.Contains(output, "bad input") {
		t.Errorf("expected message, got: %q", output)
	}
	if !strings.Contains(output, "try again") {
		t.Errorf("expected hint, got: %q", output)
	}
}

func TestHandleErrorUserErrorJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	_ = cmd.Flags().Set("json", "true")
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)

	ue := &vault.UserError{Message: "bad input", Hint: "try again"}
	handleError(cmd, ue)

	var result map[string]any
	if err := json.Unmarshal(errBuf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, errBuf.String())
	}
	if result["error"] != "bad input" {
		t.Errorf("error = %v, want 'bad input'", result["error"])
	}
	if result["hint"] != "try again" {
		t.Errorf("hint = %v, want 'try again'", result["hint"])
	}
}

func TestHandleErrorRegularError(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)

	handleError(cmd, errors.New("something broke"))

	output := errBuf.String()
	if !strings.Contains(output, "something broke") {
		t.Errorf("expected error message, got: %q", output)
	}
}

func TestHandleErrorRegularErrorJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	_ = cmd.Flags().Set("json", "true")
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)

	handleError(cmd, errors.New("broke"))

	var result map[string]any
	if err := json.Unmarshal(errBuf.Bytes(), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}
	if result["error"] != "broke" {
		t.Errorf("error = %v, want 'broke'", result["error"])
	}
}

// --- collectCommands / filterByCategory / visibleSubcommands tests ---

func TestCollectCommands(t *testing.T) {
	commands := collectCommands(rootCmd, "", "")
	if len(commands) == 0 {
		t.Error("expected non-empty commands")
	}
	// Root should have category "root"
	for _, c := range commands {
		if c.FullPath == "anvil" && c.Category != "root" {
			t.Errorf("root category = %q, want 'root'", c.Category)
		}
	}
}

func TestFilterByCategory(t *testing.T) {
	commands := []aiCommandInfo{
		{Name: "a", Category: "vault"},
		{Name: "b", Category: "env"},
		{Name: "c", Category: "vault"},
	}

	filtered := filterByCategory(commands, "vault")
	if len(filtered) != 2 {
		t.Errorf("filtered count = %d, want 2", len(filtered))
	}

	filtered = filterByCategory(commands, "VAULT") // case insensitive
	if len(filtered) != 2 {
		t.Errorf("case-insensitive filtered count = %d, want 2", len(filtered))
	}

	filtered = filterByCategory(commands, "nope")
	if len(filtered) != 0 {
		t.Errorf("nonexistent category filtered count = %d, want 0", len(filtered))
	}
}

func TestVisibleSubcommands(t *testing.T) {
	parent := &cobra.Command{Use: "parent"}
	visible := &cobra.Command{Use: "visible"}
	hidden := &cobra.Command{Use: "hidden", Hidden: true}
	helpCmd := &cobra.Command{Use: "help"}

	parent.AddCommand(visible, hidden, helpCmd)

	result := visibleSubcommands(parent)
	if len(result) != 1 {
		t.Fatalf("expected 1 visible command, got %d", len(result))
	}
	if result[0].Name() != "visible" {
		t.Errorf("visible command = %q, want 'visible'", result[0].Name())
	}
}

func TestFlattenCommands(t *testing.T) {
	parent := &cobra.Command{Use: "root", Short: "Root cmd"}
	child := &cobra.Command{Use: "child", Short: "Child cmd"}
	parent.AddCommand(child)

	entries := flattenCommands(parent, "")
	if len(entries) < 2 {
		t.Errorf("expected at least 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "root" {
		t.Errorf("first entry = %q, want 'root'", entries[0].Name)
	}
}

func TestPrintTree(t *testing.T) {
	parent := &cobra.Command{Use: "root"}
	child1 := &cobra.Command{Use: "alpha"}
	child2 := &cobra.Command{Use: "beta"}
	parent.AddCommand(child1, child2)

	var buf bytes.Buffer
	printTree(&buf, parent, "")
	output := buf.String()

	if !strings.Contains(output, "alpha") || !strings.Contains(output, "beta") {
		t.Errorf("expected alpha and beta in tree, got: %q", output)
	}
	if !strings.Contains(output, "├") || !strings.Contains(output, "└") {
		t.Errorf("expected tree connectors, got: %q", output)
	}
}

// --- sealed vault error tests ---

func TestSealedOperationsCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "K1", "V1", "-p", "test")

	// Seal
	execCmd(t, "vault", "seal")
	defer execCmd(t, "vault", "unseal")

	tests := []struct {
		name string
		args []string
	}{
		{"set", []string{"vault", "set", "K2", "V2", "-p", "test"}},
		{"list", []string{"vault", "list", "-p", "test"}},
		{"delete", []string{"vault", "delete", "K1", "-p", "test"}},
		{"export", []string{"vault", "export", "-p", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, stderr := execCmd(t, tt.args...)
			if !strings.Contains(stderr, "sealed") {
				t.Errorf("expected sealed error for %s, got: %q", tt.name, stderr)
			}
		})
	}
}

// --- UserError without hint ---

func TestHandleErrorUserErrorNoHint(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	errBuf := &bytes.Buffer{}
	cmd.SetErr(errBuf)

	ue := &vault.UserError{Message: "no hint error"}
	handleError(cmd, ue)

	output := errBuf.String()
	if !strings.Contains(output, "no hint error") {
		t.Errorf("expected message, got: %q", output)
	}
	if strings.Contains(output, "Hint") {
		t.Error("should not show Hint line when hint is empty")
	}
}
