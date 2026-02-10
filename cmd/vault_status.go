package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault status",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := vault.GetStatus(nil)
		if err != nil {
			return err
		}

		outputResult(cmd, status, func() {
			w := cmd.OutOrStdout()
			if !status.Initialized {
				_, _ = fmt.Fprintln(w, "Vault is not initialized.")
				_, _ = fmt.Fprintf(w, "Database path: %s\n", status.DBPath)
				_, _ = fmt.Fprintln(w, "Run 'profile vault init' to initialize.")
				return
			}

			_, _ = fmt.Fprintln(w, "Vault Status")
			_, _ = fmt.Fprintln(w, "  Initialized:  yes")
			_, _ = fmt.Fprintf(w, "  Database:     %s\n", status.DBPath)
			_, _ = fmt.Fprintf(w, "  Profiles:     %d\n", status.ProfileCount)
			_, _ = fmt.Fprintf(w, "  Secrets:      %d\n", status.SecretCount)
			_, _ = fmt.Fprintf(w, "  Key version:  %d\n", status.KeyVersion)
			if !status.CreatedAt.IsZero() {
				_, _ = fmt.Fprintf(w, "  Created:      %s\n", status.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		})
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultStatusCmd)
}
