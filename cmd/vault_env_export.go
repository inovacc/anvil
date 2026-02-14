package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultEnvExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export released secrets as environment variables",
	Long:  "Exports vault secrets in a format suitable for shell evaluation. Requires an active release.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		format, _ := cmd.Flags().GetString("format")

		entries, err := v.EnvExport(profileName)
		if err != nil {
			return err
		}

		w := cmd.OutOrStdout()

		switch format {
		case "json":
			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			_, _ = fmt.Fprintln(w, string(data))
		case "env":
			for _, e := range entries {
				_, _ = fmt.Fprintf(w, "%s=%s\n", e.Key, e.Value)
			}
		case "export":
			for _, e := range entries {
				_, _ = fmt.Fprintf(w, "export %s=%q\n", e.Key, e.Value)
			}
		case "powershell":
			for _, e := range entries {
				_, _ = fmt.Fprintf(w, "$env:%s=%q\n", e.Key, e.Value)
			}
		default:
			return fmt.Errorf("unsupported format: %s (use json, env, export, or powershell)", format)
		}

		return nil
	},
}

func init() {
	vaultEnvExportCmd.Flags().StringP("profile", "p", "", "Target profile (default: released profile)")
	vaultEnvExportCmd.Flags().StringP("format", "f", "env", "Output format (json, env, export, powershell)")
	vaultEnvCmd.AddCommand(vaultEnvExportCmd)
}
