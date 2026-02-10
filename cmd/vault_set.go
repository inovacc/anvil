package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set an encrypted secret",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		description, _ := cmd.Flags().GetString("description")

		if err := v.Set(args[0], args[1], profileName, description); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key     string `json:"key"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Secret %q set.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q set.\n", args[0])
		})
		return nil
	},
}

func init() {
	vaultSetCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultSetCmd.Flags().StringP("description", "d", "", "Secret description")
	vaultCmd.AddCommand(vaultSetCmd)
}
