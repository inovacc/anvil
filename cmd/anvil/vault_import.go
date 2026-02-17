package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
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

var varRefRe = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

func parseEnvData(data []byte) ([]vault.SecretEntry, error) {
	var entries []vault.SecretEntry

	vars := make(map[string]string)
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip export prefix.
		line = strings.TrimPrefix(line, "export ")

		key, rest, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)

		var value string

		switch {
		case len(rest) > 0 && rest[0] == '"':
			// Double-quoted: consume until unescaped closing quote, may span lines.
			raw := rest[1:]

			var buf strings.Builder

			closed := false

			for {
				for j := 0; j < len(raw); j++ {
					if raw[j] == '\\' && j+1 < len(raw) {
						switch raw[j+1] {
						case 'n':
							buf.WriteByte('\n')
						case 't':
							buf.WriteByte('\t')
						case '\\':
							buf.WriteByte('\\')
						case '"':
							buf.WriteByte('"')
						default:
							buf.WriteByte('\\')
							buf.WriteByte(raw[j+1])
						}

						j++

						continue
					}

					if raw[j] == '"' {
						closed = true
						break
					}

					buf.WriteByte(raw[j])
				}

				if closed {
					break
				}
				// Continue to next line.
				i++
				if i >= len(lines) {
					break
				}

				buf.WriteByte('\n')

				raw = lines[i]
			}

			value = expandVars(buf.String(), vars)
		case len(rest) > 0 && rest[0] == '\'':
			// Single-quoted: literal, no escapes, no substitution.
			end := strings.Index(rest[1:], "'")
			if end >= 0 {
				value = rest[1 : 1+end]
			} else {
				value = rest[1:]
			}
		default:
			// Unquoted: strip inline comments, expand vars.
			val := rest
			if idx := strings.Index(val, " #"); idx >= 0 {
				val = val[:idx]
			}

			value = expandVars(strings.TrimSpace(val), vars)
		}

		vars[key] = value
		entries = append(entries, vault.SecretEntry{
			Key:   key,
			Value: value,
		})
	}

	return entries, nil
}

func expandVars(s string, vars map[string]string) string {
	return varRefRe.ReplaceAllStringFunc(s, func(match string) string {
		submatch := varRefRe.FindStringSubmatch(match)

		name := submatch[1]
		if name == "" {
			name = submatch[2]
		}

		if v, ok := vars[name]; ok {
			return v
		}

		return match
	})
}

func init() {
	vaultImportCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultImportCmd.Flags().StringP("format", "f", "json", "Input format (json, env)")
	registerCommand(vaultImportCmd)
}
