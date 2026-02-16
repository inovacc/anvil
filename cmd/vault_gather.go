package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var secretPattern = regexp.MustCompile(`(?i)(password|secret|token|api_key|apikey|access_key|private_key|credential|auth)`)

var defaultExcludes = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".venv":        true,
	"__pycache__":  true,
	"dist":         true,
	"build":        true,
}

type gatheredFile struct {
	Path    string              `json:"path"`
	Secrets []vault.SecretEntry `json:"secrets"`
}

var vaultGatherCmd = &cobra.Command{
	Use:   "gather [directory]",
	Short: "Discover and import secrets from files",
	Long:  "Recursively discovers .env and config files containing secrets and imports them into a vault profile.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGather,
}

func runGather(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve directory: %w", err)
	}

	maxDepth, _ := cmd.Flags().GetInt("depth")
	extraExcludes, _ := cmd.Flags().GetStringSlice("exclude")
	profileName, _ := cmd.Flags().GetString("profile")
	autoYes, _ := cmd.Flags().GetBool("yes")

	excludes := make(map[string]bool)
	maps.Copy(excludes, defaultExcludes)
	for _, e := range extraExcludes {
		excludes[e] = true
	}

	files, err := discoverSecrets(dir, maxDepth, excludes)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		outputResult(cmd, struct {
			Message string `json:"message"`
		}{"No secret files found."}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No secret files found.")
		})
		return nil
	}

	// Dedup: last-file-wins by key.
	seen := make(map[string]vault.SecretEntry)
	var order []string
	for _, f := range files {
		for _, s := range f.Secrets {
			if _, exists := seen[s.Key]; !exists {
				order = append(order, s.Key)
			}
			seen[s.Key] = s
		}
	}

	entries := make([]vault.SecretEntry, 0, len(order))
	for _, k := range order {
		entries = append(entries, seen[k])
	}

	if !autoYes {
		reader := bufio.NewReader(os.Stdin)

		_, _ = fmt.Fprintf(os.Stderr, "Found secrets in %d file(s):\n", len(files))
		for _, f := range files {
			rel, _ := filepath.Rel(dir, f.Path)
			if rel == "" {
				rel = f.Path
			}
			_, _ = fmt.Fprintf(os.Stderr, "  %s (%d secrets)\n", rel, len(f.Secrets))
		}

		_, _ = fmt.Fprintf(os.Stderr, "\nKeys to import (%d total):\n", len(entries))
		for _, e := range entries {
			_, _ = fmt.Fprintf(os.Stderr, "  %s = ***\n", e.Key)
		}

		_, _ = fmt.Fprintf(os.Stderr, "\nCreate profile with %d secrets? [y/N]: ", len(entries))
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			_, _ = fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}

		if profileName == "" {
			defaultName := filepath.Base(dir)
			_, _ = fmt.Fprintf(os.Stderr, "Profile name [%s]: ", defaultName)
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "" {
				name = defaultName
			}
			profileName = name
		}
	} else {
		if profileName == "" {
			return fmt.Errorf("--profile is required with --yes")
		}
	}

	v, err := vault.Open(nil)
	if err != nil {
		return err
	}
	defer func() { _ = v.Close() }()

	if err := v.CreateProfile(profileName, "Gathered from "+dir, false); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}

	if err := v.Import(entries, profileName); err != nil {
		return err
	}

	jsonFiles := make([]string, 0, len(files))
	for _, f := range files {
		rel, _ := filepath.Rel(dir, f.Path)
		if rel == "" {
			rel = f.Path
		}
		jsonFiles = append(jsonFiles, rel)
	}

	outputResult(cmd, struct {
		Files   []string `json:"files"`
		Secrets []string `json:"secrets"`
		Profile string   `json:"profile"`
		Count   int      `json:"count"`
	}{
		Files:   jsonFiles,
		Secrets: order,
		Profile: profileName,
		Count:   len(entries),
	}, func() {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d secrets into profile %q.\n", len(entries), profileName)
	})

	return nil
}

func discoverSecrets(root string, maxDepth int, excludes map[string]bool) ([]gatheredFile, error) {
	var results []gatheredFile
	rootDepth := strings.Count(root, string(os.PathSeparator))

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}

		if d.IsDir() {
			if excludes[d.Name()] {
				return fs.SkipDir
			}
			depth := strings.Count(path, string(os.PathSeparator)) - rootDepth
			if depth > maxDepth {
				return fs.SkipDir
			}
			return nil
		}

		name := d.Name()

		// .env files
		if name == ".env" || strings.HasPrefix(name, ".env.") {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			entries, err := parseEnvData(data)
			if err != nil || len(entries) == 0 {
				return nil
			}
			results = append(results, gatheredFile{Path: path, Secrets: entries})
			return nil
		}

		// Config files: only extract secret-pattern keys
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".json":
			secrets := extractJSONSecrets(path)
			if len(secrets) > 0 {
				results = append(results, gatheredFile{Path: path, Secrets: secrets})
			}
		case ".yml", ".yaml":
			secrets := extractYAMLSecrets(path)
			if len(secrets) > 0 {
				results = append(results, gatheredFile{Path: path, Secrets: secrets})
			}
		}

		return nil
	})

	return results, err
}

func extractJSONSecrets(path string) []vault.SecretEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	flat := make(map[string]string)
	flattenMap("", raw, flat)
	return filterSecretKeys(flat)
}

func extractYAMLSecrets(path string) []vault.SecretEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var raw any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil
	}

	flat := make(map[string]string)
	flattenMap("", raw, flat)
	return filterSecretKeys(flat)
}

func flattenMap(prefix string, val any, out map[string]string) {
	switch v := val.(type) {
	case map[string]any:
		for k, child := range v {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			flattenMap(key, child, out)
		}
	case []any:
		for i, child := range v {
			key := fmt.Sprintf("%s.%d", prefix, i)
			flattenMap(key, child, out)
		}
	default:
		if prefix != "" && v != nil {
			out[prefix] = fmt.Sprintf("%v", v)
		}
	}
}

func filterSecretKeys(flat map[string]string) []vault.SecretEntry {
	var entries []vault.SecretEntry
	for k, v := range flat {
		if secretPattern.MatchString(k) {
			entries = append(entries, vault.SecretEntry{Key: k, Value: v})
		}
	}
	return entries
}

func init() {
	vaultGatherCmd.Flags().StringP("profile", "p", "", "Target profile name")
	vaultGatherCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompts")
	vaultGatherCmd.Flags().IntP("depth", "d", 5, "Max directory depth")
	vaultGatherCmd.Flags().StringSliceP("exclude", "e", nil, "Extra directories to exclude")
	vaultCmd.AddCommand(vaultGatherCmd)
}
