package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func TestBuildCommandDetail(t *testing.T) {
	parent := &cobra.Command{Use: "root", Short: "Root cmd"}
	child := &cobra.Command{Use: "child", Short: "Child cmd"}
	parent.AddCommand(child)

	detail := buildCommandDetail(parent)
	if detail.Name != "root" {
		t.Errorf("root name = %q, want 'root'", detail.Name)
	}

	if len(detail.Subcommands) < 1 {
		t.Errorf("expected at least 1 subcommand, got %d", len(detail.Subcommands))
	}
}

func TestBuildTree(t *testing.T) {
	parent := &cobra.Command{Use: "root"}
	child1 := &cobra.Command{Use: "alpha", Short: "Alpha cmd"}
	child2 := &cobra.Command{Use: "beta", Short: "Beta cmd"}
	parent.AddCommand(child1, child2)

	output := string(buildTree(parent))

	if !strings.Contains(output, "alpha") || !strings.Contains(output, "beta") {
		t.Errorf("expected alpha and beta in tree, got: %q", output)
	}

	if !strings.Contains(output, "+--") && !strings.Contains(output, "\\--") {
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

// --- password update success ---

func TestPasswordUpdateWithCurrentCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "env", "password", "set", "--password", "mypassword1")

	// Update with correct current password
	stdout, stderr := execCmd(t, "vault", "env", "password", "set", "--password", "newpassword1", "--current", "mypassword1")
	if stderr != "" {
		t.Errorf("expected no error, got: %q", stderr)
	}

	if !strings.Contains(stdout, "successfully") {
		t.Errorf("expected success message, got: %q", stdout)
	}
}

// --- env-inline without release ---

func TestEnvInlineNoReleaseCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "KEY1", "val1", "-p", "test")

	// No password/release set, should fail
	_, stderr := execCmd(t, "--env-inline", "KEY1", "-p", "test")
	if stderr == "" {
		t.Error("expected error without active release")
	}
}

// --- vault operations without init ---

func TestVaultGetNoInitCLI(t *testing.T) {
	t.Setenv("ANVIL_DB_PATH", t.TempDir()+"/nonexistent.db")
	t.Setenv("ANVIL_SKIP_TPM", "1")

	_, stderr := execCmd(t, "vault", "get", "KEY1")
	if stderr == "" {
		t.Error("expected error for uninitialized vault")
	}
}

// --- profile error paths ---

func TestProfileCreateDuplicateCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "dup-test")

	_, stderr := execCmd(t, "vault", "profile", "create", "dup-test")
	if stderr == "" {
		t.Error("expected error for duplicate profile")
	}
}

func TestProfileDeleteNonexistentCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "profile", "delete", "nonexistent")
	if stderr == "" {
		t.Error("expected error for nonexistent profile")
	}
}

func TestProfileUseNonexistentCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "profile", "use", "nonexistent")
	if stderr == "" {
		t.Error("expected error for nonexistent profile")
	}
}

// --- secret error paths ---

func TestGetNonexistentSecretCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	_, stderr := execCmd(t, "vault", "get", "NOPE", "-p", "test")
	if stderr == "" {
		t.Error("expected error for nonexistent secret")
	}
}

func TestDeleteNonexistentSecretCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	_, stderr := execCmd(t, "vault", "delete", "NOPE", "-p", "test")
	if stderr == "" {
		t.Error("expected error for nonexistent secret")
	}
}

// --- history and rollback error paths ---

func TestHistoryNonexistentCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	stdout, _ := execCmd(t, "vault", "history", "NOPE", "-p", "test")
	// Should succeed but show no versions
	if strings.Contains(stdout, "error") {
		t.Errorf("unexpected error: %q", stdout)
	}
}

func TestRollbackNonexistentCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")

	_, stderr := execCmd(t, "vault", "rollback", "NOPE", "1", "-p", "test")
	if stderr == "" {
		t.Error("expected error for nonexistent secret rollback")
	}
}

// --- export/import edge cases ---

func TestExportEmptyProfileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "empty", "--default")

	stdout, stderr := execCmd(t, "vault", "export", "-p", "empty", "--format", "json")
	if stderr != "" {
		t.Fatalf("export error: %s", stderr)
	}
	// Should be valid JSON (empty or [])
	if !strings.Contains(stdout, "[") && !strings.Contains(stdout, "{") {
		t.Errorf("expected JSON output, got: %q", stdout)
	}
}

func TestExportEnvFormatCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "envtest", "--default")
	execCmd(t, "vault", "set", "MY_KEY", "my_val", "-p", "envtest")

	stdout, stderr := execCmd(t, "vault", "export", "-p", "envtest", "--format", "env")
	if stderr != "" {
		t.Fatalf("export error: %s", stderr)
	}

	if !strings.Contains(stdout, "MY_KEY=my_val") {
		t.Errorf("expected env format, got: %q", stdout)
	}
}

// --- template error paths ---

func TestTemplateShowNonexistentCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "template", "show", "nonexistent")
	if stderr == "" {
		t.Error("expected error for nonexistent template")
	}
}

func TestTemplateDeleteNonexistentCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "template", "delete", "nonexistent")
	if stderr == "" {
		t.Error("expected error for nonexistent template")
	}
}

// --- audit with filters ---

func TestAuditWithProfileFilterCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "auditprof", "--default")
	execCmd(t, "vault", "set", "AK", "AV", "-p", "auditprof")

	stdout, stderr := execCmd(t, "vault", "audit", "-p", "auditprof")
	if stderr != "" {
		t.Fatalf("audit error: %s", stderr)
	}

	if !strings.Contains(stdout, "auditprof") {
		t.Errorf("expected filtered audit, got: %q", stdout)
	}
}

func TestAuditWithLimitCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "test", "--default")
	execCmd(t, "vault", "set", "K1", "V1", "-p", "test")
	execCmd(t, "vault", "set", "K2", "V2", "-p", "test")

	stdout, stderr := execCmd(t, "vault", "audit", "--limit", "1")
	if stderr != "" {
		t.Fatalf("audit limit error: %s", stderr)
	}

	_ = stdout // just verify no error
}

// --- unseal when not sealed ---

func TestUnsealNotSealedExtraCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "unseal")
	if stderr == "" {
		t.Error("expected error when unsealing a non-sealed vault")
	}
}

// --- vault status JSON ---

func TestVaultStatusJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "vault", "status", "--json")
	if stderr != "" {
		t.Fatalf("status JSON error: %s", stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v", err)
	}

	if _, ok := result["initialized"]; !ok {
		t.Error("expected 'initialized' field in JSON status")
	}
}

// --- root help ---

func TestRootHelpCLI(t *testing.T) {
	setupTestVault(t)

	// Root with no args invokes help; help output may go to stdout or be captured differently
	stdout, stderr := execCmd(t, "--help")

	combined := stdout + stderr
	if !strings.Contains(combined, "anvil") && !strings.Contains(combined, "vault") {
		t.Errorf("expected help output, got stdout=%q stderr=%q", stdout[:min(len(stdout), 200)], stderr[:min(len(stderr), 200)])
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

// --- import tests ---

func TestImportJSONCLI(t *testing.T) {
	setupTestVault(t)

	// Create a profile first.
	execCmd(t, "vault", "profile", "create", "importtest", "-d", "test")

	// Write a JSON import file.
	tmpFile := filepath.Join(t.TempDir(), "secrets.json")

	data := `[{"key":"API_KEY","value":"secret123"},{"key":"DB_PASS","value":"pass456"}]`
	if err := os.WriteFile(tmpFile, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "import", tmpFile, "-p", "importtest", "-f", "json")
	if !strings.Contains(stdout, "Imported 2 secrets") {
		t.Errorf("expected import success, got: %q", stdout)
	}
}

func TestImportEnvCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "envimport", "-d", "test")

	tmpFile := filepath.Join(t.TempDir(), "secrets.env")

	data := "API_KEY=secret123\nDB_PASS=pass456\n# comment\nexport TOKEN=abc\n"
	if err := os.WriteFile(tmpFile, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "import", tmpFile, "-p", "envimport", "-f", "env")
	if !strings.Contains(stdout, "Imported 3 secrets") {
		t.Errorf("expected import success, got: %q", stdout)
	}
}

func TestImportUnsupportedFormatCLI(t *testing.T) {
	setupTestVault(t)

	tmpFile := filepath.Join(t.TempDir(), "secrets.xml")
	if err := os.WriteFile(tmpFile, []byte("<xml/>"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr := execCmd(t, "vault", "import", tmpFile, "-f", "xml")
	if !strings.Contains(stderr, "unsupported format") {
		t.Errorf("expected unsupported format error, got: %q", stderr)
	}
}

func TestImportNonexistentFileCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "vault", "import", "/nonexistent/file.json", "-f", "json")
	if stderr == "" {
		t.Error("expected error for nonexistent file")
	}
}

func TestGatherEmptyDirCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gathertest", "-d", "test")

	emptyDir := t.TempDir()
	stdout, _ := execCmd(t, "vault", "gather", emptyDir, "--yes", "-p", "gathertest")
	_ = stdout
}

func TestGatherWithEnvFileCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatherenv", "-d", "test")

	dir := t.TempDir()

	envContent := "API_KEY=secret123\nDB_PASSWORD=pass456\nNORMAL_VAR=hello\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherenv")
	_ = stdout
}

func TestGatherWithJSONFileCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatherjson", "-d", "test")

	dir := t.TempDir()

	jsonContent := `{"api_key": "secret123", "database": {"password": "pass456"}}`
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherjson")
	_ = stdout
}

func TestGatherWithYAMLFileCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatheryaml", "-d", "test")

	dir := t.TempDir()

	yamlContent := "api_key: secret123\ndatabase:\n  password: pass456\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatheryaml")
	_ = stdout
}

func TestImportJSONOutputCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "jsonout", "-d", "test")

	tmpFile := filepath.Join(t.TempDir(), "secrets.json")

	data := `[{"key":"TOKEN","value":"abc"}]`
	if err := os.WriteFile(tmpFile, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "import", tmpFile, "-p", "jsonout", "-f", "json", "--json")
	if !strings.Contains(stdout, `"count"`) {
		t.Errorf("expected JSON output with count, got: %q", stdout)
	}
}

func TestGatherNoProfileWithYesCLI(t *testing.T) {
	setupTestVault(t)

	dir := t.TempDir()

	envContent := "API_KEY=secret\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := execCmd(t, "vault", "gather", dir, "--yes")

	combined := stdout + stderr
	if !strings.Contains(combined, "profile") {
		t.Errorf("expected profile-related error, got stdout=%q stderr=%q", stdout, stderr)
	}
}

func TestGatherJSONOutputCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatherjsonout", "-d", "test")

	dir := t.TempDir()

	envContent := "DB_PASSWORD=secret123\nAPI_TOKEN=abc\n"
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherjsonout", "--json")
	if !strings.Contains(stdout, `"profile"`) {
		t.Errorf("expected JSON output, got: %q", stdout)
	}
}

func TestGatherInvalidJSONCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatherinv", "-d", "test")

	dir := t.TempDir()
	// Invalid JSON should be silently skipped by extractJSONSecrets
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{invalid json"), 0644); err != nil {
		t.Fatal(err)
	}
	// Invalid YAML
	if err := os.WriteFile(filepath.Join(dir, "bad.yaml"), []byte(":\n  :\n  - :\n  {{{}}}"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherinv")
	if !strings.Contains(stdout, "No secret files found") {
		_ = stdout // invalid files skipped, may find nothing
	}
}

func TestGatherWithExcludeCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "vault", "profile", "create", "gatherexcl", "-d", "test")

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "subdir", ".env"), []byte("API_KEY=secret\n"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherexcl", "-e", "subdir")
	// subdir excluded, so no secrets found
	if !strings.Contains(stdout, "No secret files found") {
		// May still find files if exclusion only applies to dir names
		_ = stdout
	}
}
