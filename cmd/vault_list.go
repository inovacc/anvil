package cmd

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
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

			tw := tableWriter(w)
			_, _ = fmt.Fprintln(tw, "KEY\tDESCRIPTION\tCREATED\tUPDATED")
			for _, s := range secrets {
				updated := "-"
				if !s.UpdatedAt.IsZero() {
					updated = s.UpdatedAt.Format("2006-01-02 15:04:05")
				}
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
					s.Key,
					s.Description,
					s.CreatedAt.Format("2006-01-02 15:04:05"),
					updated,
				)
			}
			_ = tw.Flush()
		})
		return nil
	},
}

func init() {
	vaultListCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	registerCommand(vaultListCmd)
}
