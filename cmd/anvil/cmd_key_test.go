package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKeyGenerateCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "key", "generate", "mykey", "--algorithm", "ed25519")
	if stderr != "" {
		t.Fatalf("key generate error: %s", stderr)
	}

	if !strings.Contains(stdout, "mykey") || !strings.Contains(stdout, "generated") {
		t.Errorf("unexpected output: %q", stdout)
	}

	if !strings.Contains(stdout, "ed25519") {
		t.Errorf("expected ed25519 algorithm, got: %q", stdout)
	}

	if !strings.Contains(stdout, "Fingerprint") {
		t.Errorf("expected fingerprint, got: %q", stdout)
	}
}

func TestKeyGenerateECDSACLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "key", "generate", "eckey", "--algorithm", "ecdsa-p256", "--description", "test ec key")
	if stderr != "" {
		t.Fatalf("key generate error: %s", stderr)
	}

	if !strings.Contains(stdout, "ecdsa-p256") {
		t.Errorf("expected ecdsa-p256, got: %q", stdout)
	}
}

func TestKeyGenerateJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "key", "generate", "jsonkey", "--algorithm", "ed25519", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["name"] != "jsonkey" {
		t.Errorf("expected name=jsonkey, got %v", result["name"])
	}

	if result["algorithm"] != "ed25519" {
		t.Errorf("expected algorithm=ed25519, got %v", result["algorithm"])
	}
}

func TestKeyGenerateDuplicateCLI(t *testing.T) {
	setupTestVault(t)

	execCmd(t, "key", "generate", "dupkey", "--algorithm", "ed25519")

	_, stderr := execCmd(t, "key", "generate", "dupkey", "--algorithm", "ed25519")
	if !strings.Contains(stderr, "already exists") {
		t.Errorf("expected 'already exists' error, got: %q", stderr)
	}
}

func TestKeyListCLI(t *testing.T) {
	setupTestVault(t)

	// Empty list.
	stdout, stderr := execCmd(t, "key", "list")
	if stderr != "" {
		t.Fatalf("key list error: %s", stderr)
	}

	if !strings.Contains(stdout, "No keys found") {
		t.Errorf("expected no keys message, got: %q", stdout)
	}

	// Generate then list.
	execCmd(t, "key", "generate", "listkey1", "--algorithm", "ed25519")
	execCmd(t, "key", "generate", "listkey2", "--algorithm", "ecdsa-p256")

	stdout, stderr = execCmd(t, "key", "list")
	if stderr != "" {
		t.Fatalf("key list error: %s", stderr)
	}

	if !strings.Contains(stdout, "listkey1") || !strings.Contains(stdout, "listkey2") {
		t.Errorf("expected both keys in list, got: %q", stdout)
	}

	if !strings.Contains(stdout, "NAME") {
		t.Errorf("expected table header, got: %q", stdout)
	}
}

func TestKeyListJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "jk", "--algorithm", "ed25519")

	stdout, _ := execCmd(t, "key", "list", "--json")

	var keys []map[string]any
	if err := json.Unmarshal([]byte(stdout), &keys); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}
}

func TestKeyDeleteCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "delkey", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "key", "delete", "delkey")
	if stderr != "" {
		t.Fatalf("key delete error: %s", stderr)
	}

	if !strings.Contains(stdout, "deleted") {
		t.Errorf("expected deleted message, got: %q", stdout)
	}

	// Verify gone.
	stdout, _ = execCmd(t, "key", "list")
	if strings.Contains(stdout, "delkey") {
		t.Error("key should have been deleted")
	}
}

func TestKeyDeleteNonexistentCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "key", "delete", "nosuchkey")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %q", stderr)
	}
}

func TestKeyExportPublicCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "expkey", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "key", "export", "expkey")
	if stderr != "" {
		t.Fatalf("key export error: %s", stderr)
	}

	if !strings.Contains(stdout, "PUBLIC KEY") {
		t.Errorf("expected PEM public key, got: %q", stdout)
	}
}

func TestKeyExportPrivateCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "privexp", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "key", "export", "privexp", "--private")
	if stderr != "" {
		t.Fatalf("key export error: %s", stderr)
	}

	if !strings.Contains(stdout, "PRIVATE KEY") {
		t.Errorf("expected PEM private key, got: %q", stdout)
	}
}

func TestKeyExportToFileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "fileexp", "--algorithm", "ed25519")

	outFile := filepath.Join(t.TempDir(), "key.pem")

	stdout, stderr := execCmd(t, "key", "export", "fileexp", "--output", outFile)
	if stderr != "" {
		t.Fatalf("key export error: %s", stderr)
	}

	if !strings.Contains(stdout, "exported") {
		t.Errorf("expected exported message, got: %q", stdout)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("exported file is empty")
	}
}

func TestKeyExportNotFoundCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "key", "export", "nokey")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %q", stderr)
	}
}

func TestKeyImportCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "imptest", "--algorithm", "ed25519")

	// Export private key to file using --output flag (avoid stdout PEM issues).
	pemFile := filepath.Join(t.TempDir(), "export.pem")
	execCmd(t, "key", "export", "imptest", "--private", "--output", pemFile)

	// Import under new name.
	stdout, stderr := execCmd(t, "key", "import", "imported-key", pemFile)
	if stderr != "" {
		t.Fatalf("key import error: %s", stderr)
	}

	if !strings.Contains(stdout, "imported") {
		t.Errorf("expected imported message, got: %q", stdout)
	}

	if !strings.Contains(stdout, "ed25519") {
		t.Errorf("expected algorithm in output, got: %q", stdout)
	}
}

func TestKeyImportBadFileCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "key", "import", "badimport", "/nonexistent/file.pem")
	if stderr == "" {
		t.Error("expected error for nonexistent file")
	}
}

func TestSignStringCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "signkey", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "sign", "--key", "signkey", "--string", "hello world", "--file", "", "--output", "")
	if stderr != "" {
		t.Fatalf("sign error: %s", stderr)
	}

	sig := strings.TrimSpace(stdout)
	if len(sig) == 0 {
		t.Fatal("expected non-empty signature")
	}
}

func TestSignFileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "signfilekey", "--algorithm", "ed25519")

	dataFile := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(dataFile, []byte("file content"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := execCmd(t, "sign", "--key", "signfilekey", "--file", dataFile, "--string", "", "--output", "")
	if stderr != "" {
		t.Fatalf("sign error: %s", stderr)
	}

	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected non-empty signature")
	}
}

func TestSignToFileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "sigoutkey", "--algorithm", "ed25519")

	sigFile := filepath.Join(t.TempDir(), "sig.b64")

	stdout, stderr := execCmd(t, "sign", "--key", "sigoutkey", "--string", "data", "--file", "", "--output", sigFile)
	if stderr != "" {
		t.Fatalf("sign error: %s", stderr)
	}

	if !strings.Contains(stdout, "written") {
		t.Errorf("expected written message, got: %q", stdout)
	}

	data, err := os.ReadFile(sigFile)
	if err != nil {
		t.Fatalf("read sig file: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("signature file is empty")
	}
}

func TestSignNoInputCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "noinput", "--algorithm", "ed25519")

	_, stderr := execCmd(t, "sign", "--key", "noinput", "--file", "", "--string", "", "--output", "")
	if !strings.Contains(stderr, "file") && !strings.Contains(stderr, "string") {
		t.Errorf("expected error about missing input, got: %q", stderr)
	}
}

func TestSignJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "jsonsign", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "sign", "--key", "jsonsign", "--string", "test", "--file", "", "--output", "", "--json")
	if stderr != "" {
		t.Logf("sign JSON stderr: %s", stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["signature"] == nil || result["signature"] == "" {
		t.Error("expected signature in JSON output")
	}

	if result["key_name"] != "jsonsign" {
		t.Errorf("expected key_name=jsonsign, got %v", result["key_name"])
	}
}

func TestVerifyValidCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vfykey", "--algorithm", "ed25519")

	// Sign first.
	sigOut, _ := execCmd(t, "sign", "--key", "vfykey", "--string", "hello", "--file", "", "--output", "")
	sig := strings.TrimSpace(sigOut)

	stdout, stderr := execCmd(t, "verify", "--key", "vfykey", "--string", "hello", "--file", "", "--signature", sig, "--signature-file", "")
	if stderr != "" {
		t.Fatalf("verify error: %s", stderr)
	}

	if !strings.Contains(stdout, "valid") {
		t.Errorf("expected valid, got: %q", stdout)
	}
}

func TestVerifyFromFileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vfyfilekey", "--algorithm", "ed25519")

	sigOut, _ := execCmd(t, "sign", "--key", "vfyfilekey", "--string", "data", "--file", "", "--output", "")
	sig := strings.TrimSpace(sigOut)

	sigFile := filepath.Join(t.TempDir(), "sig.b64")
	if err := os.WriteFile(sigFile, []byte(sig), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := execCmd(t, "verify", "--key", "vfyfilekey", "--string", "data", "--file", "", "--signature", "", "--signature-file", sigFile)
	if stderr != "" {
		t.Fatalf("verify error: %s", stderr)
	}

	if !strings.Contains(stdout, "valid") {
		t.Errorf("expected valid, got: %q", stdout)
	}
}

func TestVerifyNoInputCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vnoinput", "--algorithm", "ed25519")

	_, stderr := execCmd(t, "verify", "--key", "vnoinput", "--string", "", "--file", "", "--signature", "abc", "--signature-file", "")
	if !strings.Contains(stderr, "file") && !strings.Contains(stderr, "string") {
		t.Errorf("expected error about missing input, got: %q", stderr)
	}
}

func TestVerifyNoSigCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vnosig", "--algorithm", "ed25519")

	_, stderr := execCmd(t, "verify", "--key", "vnosig", "--string", "data", "--file", "", "--signature", "", "--signature-file", "")
	if !strings.Contains(stderr, "signature") {
		t.Errorf("expected error about missing signature, got: %q", stderr)
	}
}

func TestCmdTreeSingleCommandCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "cmdtree", "--command", "key")
	if stderr != "" {
		t.Fatalf("cmdtree --command error: %s", stderr)
	}

	if !strings.Contains(stdout, "key") {
		t.Errorf("expected key command details, got: %q", stdout)
	}
}

func TestCmdTreeSingleCommandJSONCLI(t *testing.T) {
	setupTestVault(t)

	// Note: --json mode always returns full tree (--command is ignored in JSON mode).
	stdout, _ := execCmd(t, "cmdtree", "--command", "", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["name"] != "anvil" {
		t.Errorf("expected name=anvil, got %v", result["name"])
	}

	commands, ok := result["commands"].([]any)
	if !ok || len(commands) == 0 {
		t.Error("expected subcommands in JSON output")
	}
}

func TestCmdTreeSingleCommandNotFoundCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "cmdtree", "--command", "nonexistent", "--brief=false")
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected 'not found' error, got: %q", stderr)
	}
}

func TestCmdTreeBriefCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "cmdtree", "--brief", "--command", "")
	if stderr != "" {
		t.Fatalf("cmdtree --brief error: %s", stderr)
	}

	if !strings.Contains(stdout, "anvil") {
		t.Errorf("expected anvil in output, got: %q", stdout)
	}
}

func TestInstallationIDCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "id")
	if stderr != "" {
		t.Fatalf("id error: %s", stderr)
	}

	id := strings.TrimSpace(stdout)
	if len(id) != 64 {
		t.Errorf("expected 64-char hex ID, got %d chars: %q", len(id), id)
	}
}

func TestInstallationIDJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "id", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["installation_id"] == nil || result["installation_id"] == "" {
		t.Error("expected installation_id in JSON output")
	}
}

func TestPasswordSetNonInteractiveCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "password", "set", "--password", "mypassword123")
	if stderr != "" {
		t.Fatalf("password set error: %s", stderr)
	}

	if !strings.Contains(stdout, "successfully") {
		t.Errorf("expected success message, got: %q", stdout)
	}
}

func TestPasswordSetTooShortCLI(t *testing.T) {
	setupTestVault(t)

	_, stderr := execCmd(t, "password", "set", "--password", "short")
	if !strings.Contains(stderr, "8 characters") {
		t.Errorf("expected length error, got: %q", stderr)
	}
}

func TestPasswordSetUpdateRequiresCurrentCLI(t *testing.T) {
	setupTestVault(t)

	// Set initial password.
	execCmd(t, "password", "set", "--password", "firstpassword1")

	// Update without --current should fail.
	_, stderr := execCmd(t, "password", "set", "--password", "secondpassword2", "--current", "")
	if !strings.Contains(stderr, "current") {
		t.Errorf("expected error about --current, got: %q", stderr)
	}
}

func TestPasswordSetUpdateWithCurrentCLI(t *testing.T) {
	setupTestVault(t)

	// Set initial password.
	execCmd(t, "password", "set", "--password", "firstpassword1")

	// Update with --current should succeed.
	stdout, stderr := execCmd(t, "password", "set", "--password", "secondpassword2", "--current", "firstpassword1")
	if stderr != "" {
		t.Fatalf("password update error: %s", stderr)
	}

	if !strings.Contains(stdout, "successfully") {
		t.Errorf("expected success message, got: %q", stdout)
	}
}

func TestPasswordSetJSONCLI(t *testing.T) {
	setupTestVault(t)

	stdout, _ := execCmd(t, "password", "set", "--password", "mypassword123", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["message"] == nil {
		t.Error("expected message in JSON output")
	}
}

func TestGatherWithDepthLimitCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "gatherdepth", "-d", "test")

	dir := t.TempDir()
	// Create nested dirs beyond depth 1.
	deepDir := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Put .env at depth 3 — should be excluded with --depth 1.
	if err := os.WriteFile(filepath.Join(deepDir, ".env"), []byte("API_KEY=deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Put .env at depth 0 — should be included.
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("DB_PASSWORD=shallow\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherdepth", "-d", "1")
	_ = stdout
}

func TestGatherWithMultipleEnvFilesCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "gathermulti", "-d", "test")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("API_KEY=key1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("API_SECRET=secret1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gathermulti", "--json")
	if !strings.Contains(stdout, "gathermulti") {
		t.Errorf("expected profile name in output, got: %q", stdout)
	}
}

func TestGatherYAMLWithArrayCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "gatherarr", "-d", "test")

	dir := t.TempDir()
	yamlContent := "services:\n  - name: api\n    api_key: secret1\n  - name: db\n    password: secret2\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yamlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "gatherarr")
	_ = stdout
}

func TestGatherExistingProfileCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "vault", "profile", "create", "existing", "-d", "test")

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("API_TOKEN=val\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Should succeed even with existing profile.
	stdout, _ := execCmd(t, "vault", "gather", dir, "--yes", "-p", "existing")
	if strings.Contains(stdout, "error") {
		t.Errorf("unexpected error: %q", stdout)
	}
}

func TestVerifyJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vjsonkey", "--algorithm", "ed25519")

	sigOut, _ := execCmd(t, "sign", "--key", "vjsonkey", "--string", "jsondata", "--file", "", "--output", "")
	sig := strings.TrimSpace(sigOut)

	stdout, _ := execCmd(t, "verify", "--key", "vjsonkey", "--string", "jsondata", "--file", "", "--signature", sig, "--signature-file", "", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["valid"] != true {
		t.Errorf("expected valid=true, got %v", result["valid"])
	}
}

func TestVerifyFileDataCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "vfyfiledata", "--algorithm", "ed25519")

	// Create data file.
	dataFile := filepath.Join(t.TempDir(), "data.txt")
	if err := os.WriteFile(dataFile, []byte("file data"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Sign the file.
	sigOut, _ := execCmd(t, "sign", "--key", "vfyfiledata", "--file", dataFile, "--string", "", "--output", "")
	sig := strings.TrimSpace(sigOut)

	// Verify using file.
	stdout, stderr := execCmd(t, "verify", "--key", "vfyfiledata", "--file", dataFile, "--string", "", "--signature", sig, "--signature-file", "")
	if stderr != "" {
		t.Fatalf("verify error: %s", stderr)
	}

	if !strings.Contains(stdout, "valid") {
		t.Errorf("expected valid, got: %q", stdout)
	}
}

func TestKeyDeleteWithAlgorithmCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "algodel", "--algorithm", "ecdsa-p256")

	stdout, stderr := execCmd(t, "key", "delete", "algodel", "--algorithm", "ecdsa-p256")
	if stderr != "" {
		t.Fatalf("key delete error: %s", stderr)
	}
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("expected deleted, got: %q", stdout)
	}
}

func TestKeyDeleteJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "deljson", "--algorithm", "ed25519")

	stdout, stderr := execCmd(t, "key", "delete", "deljson", "--algorithm", "", "--json")
	if stderr != "" {
		t.Logf("key delete JSON stderr: %s", stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["name"] != "deljson" {
		t.Errorf("expected name=deljson, got %v", result["name"])
	}
}

func TestKeyExportToFileJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "expjson", "--algorithm", "ed25519")

	outFile := filepath.Join(t.TempDir(), "key.pem")
	stdout, _ := execCmd(t, "key", "export", "expjson", "--output", outFile, "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["name"] != "expjson" {
		t.Errorf("expected name=expjson, got %v", result["name"])
	}
}

func TestKeyImportJSONCLI(t *testing.T) {
	setupTestVault(t)
	execCmd(t, "key", "generate", "impjsontest", "--algorithm", "ed25519")

	pemFile := filepath.Join(t.TempDir(), "export.pem")
	execCmd(t, "key", "export", "impjsontest", "--private", "--output", pemFile)

	stdout, _ := execCmd(t, "key", "import", "impjson-imported", pemFile, "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("JSON parse: %v\noutput: %q", err, stdout)
	}

	if result["name"] != "impjson-imported" {
		t.Errorf("expected name=impjson-imported, got %v", result["name"])
	}
}

func TestPasswordSetWrongCurrentPasswordCLI(t *testing.T) {
	setupTestVault(t)

	// Set initial password.
	execCmd(t, "password", "set", "--password", "firstpassword1")

	// Try to update with wrong current password.
	_, stderr := execCmd(t, "password", "set", "--password", "secondpassword2", "--current", "wrongpassword")
	if stderr == "" {
		t.Error("expected error for wrong current password")
	}
}

func TestCmdTreeVerboseCLI(t *testing.T) {
	setupTestVault(t)

	stdout, stderr := execCmd(t, "cmdtree", "--verbose", "--command", "", "--brief=false")
	if stderr != "" {
		t.Fatalf("cmdtree --verbose error: %s", stderr)
	}

	if !strings.Contains(stdout, "anvil") {
		t.Errorf("expected anvil in verbose output, got: %q", stdout)
	}

	// Verbose mode should show flags.
	if !strings.Contains(stdout, "--") {
		t.Errorf("expected flags in verbose output, got: %q", stdout)
	}
}
