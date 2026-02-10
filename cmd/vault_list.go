package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultListCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets in a profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		secrets, err := v.List(profileName)
		if err != nil {
			return err
		}

		outputResult(cmd, secrets, func() {
			w := cmd.OutOrStdout()
			if len(secrets) == 0 {
				_, _ = fmt.Fprintln(w, "No secrets found.")
				return
			}

			for _, s := range secrets {
				desc := ""
				if s.Description != "" {
					desc = fmt.Sprintf(" - %s", s.Description)
				}
				_, _ = fmt.Fprintf(w, "  %s%s\n", s.Key, desc)
			}
		})
		return nil
	},
}

func init() {
	vaultListCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultCmd.AddCommand(vaultListCmd)
}
