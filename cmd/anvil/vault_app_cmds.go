package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var appRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new isolated app vault",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		serviceID, _ := cmd.Flags().GetString("service-id")

		info, err := v.RegisterApp(name, description, serviceID)
		if err != nil {
			return err
		}

		outputResult(cmd, info, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "App %q registered (uuid: %s)\n", info.Name, info.UUID)
		})

		return nil
	},
}

var appListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		apps, err := v.ListApps()
		if err != nil {
			return err
		}

		outputResult(cmd, apps, func() {
			w := tableWriter(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(w, "NAME\tUUID\tSERVICE ID\tSTATUS\tSECRETS\tLAST ACCESSED")

			for _, app := range apps {
				lastAccessed := "-"
				if !app.LastAccessedAt.IsZero() {
					lastAccessed = app.LastAccessedAt.Format("2006-01-02 15:04")
				}

				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
					app.Name, app.UUID, app.ServiceID, app.Status, app.SecretCount, lastAccessed)
			}

			_ = w.Flush()
		})

		return nil
	},
}

var appInfoCmd = &cobra.Command{
	Use:   "info <name-or-uuid>",
	Short: "Show app details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		info, err := v.GetApp(args[0])
		if err != nil {
			return err
		}

		outputResult(cmd, info, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(w, "Name:          %s\n", info.Name)
			_, _ = fmt.Fprintf(w, "UUID:          %s\n", info.UUID)
			_, _ = fmt.Fprintf(w, "Description:   %s\n", info.Description)
			_, _ = fmt.Fprintf(w, "Service ID:    %s\n", info.ServiceID)
			_, _ = fmt.Fprintf(w, "Status:        %s\n", info.Status)
			_, _ = fmt.Fprintf(w, "Secrets:       %d\n", info.SecretCount)
			_, _ = fmt.Fprintf(w, "DB Path:       %s\n", info.DBPath)
			_, _ = fmt.Fprintf(w, "Created:       %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))

			if !info.LastAccessedAt.IsZero() {
				_, _ = fmt.Fprintf(w, "Last Accessed: %s\n", info.LastAccessedAt.Format("2006-01-02 15:04:05"))
			}
		})

		return nil
	},
}

var appRemoveCmd = &cobra.Command{
	Use:   "remove <name-or-uuid>",
	Short: "Remove an app and its database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		if err := v.RemoveApp(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{fmt.Sprintf("App %q removed.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "App %q removed.\n", args[0])
		})

		return nil
	},
}

var appDisableCmd = &cobra.Command{
	Use:   "disable <name-or-uuid>",
	Short: "Disable an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		if err := v.DisableApp(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{fmt.Sprintf("App %q disabled.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "App %q disabled.\n", args[0])
		})

		return nil
	},
}

var appEnableCmd = &cobra.Command{
	Use:   "enable <name-or-uuid>",
	Short: "Enable an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		if err := v.EnableApp(args[0]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{fmt.Sprintf("App %q enabled.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "App %q enabled.\n", args[0])
		})

		return nil
	},
}

var appSetCmd = &cobra.Command{
	Use:   "set <name-or-uuid> <key> <value>",
	Short: "Set a secret in an app vault",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		description, _ := cmd.Flags().GetString("description")

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		if err := av.Set(args[1], args[2], description); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key     string `json:"key"`
			Message string `json:"message"`
		}{args[1], fmt.Sprintf("Secret %q set in app %q.", args[1], args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q set in app %q.\n", args[1], args[0])
		})

		return nil
	},
}

var appGetCmd = &cobra.Command{
	Use:   "get <name-or-uuid> <key>",
	Short: "Get a secret from an app vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		value, err := av.Get(args[1])
		if err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{args[1], value}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)
		})

		return nil
	},
}

var appDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-uuid> <key>",
	Short: "Delete a secret from an app vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		if err := av.Delete(args[1]); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key     string `json:"key"`
			Message string `json:"message"`
		}{args[1], fmt.Sprintf("Secret %q deleted from app %q.", args[1], args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q deleted from app %q.\n", args[1], args[0])
		})

		return nil
	},
}

var appListSecretsCmd = &cobra.Command{
	Use:   "list-secrets <name-or-uuid>",
	Short: "List secrets in an app vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		secrets, err := av.List()
		if err != nil {
			return err
		}

		outputResult(cmd, secrets, func() {
			w := tableWriter(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(w, "KEY\tDESCRIPTION\tCREATED\tUPDATED")

			for _, sec := range secrets {
				updated := "-"
				if !sec.UpdatedAt.IsZero() {
					updated = sec.UpdatedAt.Format("2006-01-02 15:04")
				}

				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
					sec.Key, sec.Description, sec.CreatedAt.Format("2006-01-02 15:04"), updated)
			}

			_ = w.Flush()
		})

		return nil
	},
}

var appExportCmd = &cobra.Command{
	Use:   "export <name-or-uuid>",
	Short: "Export secrets from an app vault",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		entries, err := av.Export()
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		default: // env
			for _, entry := range entries {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s=%s\n", entry.Key, entry.Value)
			}
		}

		return nil
	},
}

var appImportCmd = &cobra.Command{
	Use:   "import <name-or-uuid> <file>",
	Short: "Import secrets into an app vault",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		av, err := v.OpenApp(args[0])
		if err != nil {
			return err
		}

		defer func() { _ = av.Close() }()

		format, _ := cmd.Flags().GetString("format")
		filePath := args[1]

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		var entries []vault.SecretEntry

		switch format {
		case "json":
			if err := json.Unmarshal(data, &entries); err != nil {
				return fmt.Errorf("parse JSON: %w", err)
			}
		default: // env
			scanner := bufio.NewScanner(strings.NewReader(string(data)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}

				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}

				entries = append(entries, vault.SecretEntry{
					Key:   strings.TrimSpace(parts[0]),
					Value: strings.TrimSpace(parts[1]),
				})
			}
		}

		if err := av.Import(entries); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Count   int    `json:"count"`
			Message string `json:"message"`
		}{len(entries), fmt.Sprintf("Imported %d secrets into app %q.", len(entries), args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Imported %d secrets into app %q.\n", len(entries), args[0])
		})

		return nil
	},
}

func init() {
	appRegisterCmd.Flags().String("name", "", "App name (required)")
	appRegisterCmd.Flags().String("description", "", "App description")
	appRegisterCmd.Flags().String("service-id", "", "Service identifier")
	_ = appRegisterCmd.MarkFlagRequired("name")

	appSetCmd.Flags().StringP("description", "d", "", "Secret description")

	appExportCmd.Flags().String("format", "env", "Export format (json|env)")
	appImportCmd.Flags().String("format", "env", "Import format (json|env)")

	appCmd.AddCommand(
		appRegisterCmd,
		appListCmd,
		appInfoCmd,
		appRemoveCmd,
		appDisableCmd,
		appEnableCmd,
		appSetCmd,
		appGetCmd,
		appDeleteCmd,
		appListSecretsCmd,
		appExportCmd,
		appImportCmd,
	)
}
