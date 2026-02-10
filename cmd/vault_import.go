package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import secrets from a file",
	Long:  "Imports secrets from a JSON or env file into a profile.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		format, _ := cmd.Flags().GetString("format")

		entries, err := parseImportFile(args[0], format)
		if err != nil {
			return err
		}

		if err := v.Import(entries, profileName); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Count   int    `json:"count"`
			Message string `json:"message"`
		}{len(entries), fmt.Sprintf("Imported %d secrets.", len(entries))}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d secrets.\n", len(entries))
		})
		return nil
	},
}

func parseImportFile(path, format string) ([]vault.SecretEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	switch format {
	case "json":
		var entries []vault.SecretEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}

		return entries, nil
	case "env":
		return parseEnvData(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s (use json or env)", format)
	}
}

func parseEnvData(data []byte) ([]vault.SecretEntry, error) {
	var entries []vault.SecretEntry

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		entries = append(entries, vault.SecretEntry{
			Key:   strings.TrimSpace(key),
			Value: strings.TrimSpace(value),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env: %w", err)
	}

	return entries, nil
}

func init() {
	vaultImportCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultImportCmd.Flags().StringP("format", "f", "json", "Input format (json, env)")
	vaultCmd.AddCommand(vaultImportCmd)
}
