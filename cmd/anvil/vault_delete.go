package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		if err := v.Delete(args[0], profileName); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key     string `json:"key"`
			Message string `json:"message"`
		}{args[0], fmt.Sprintf("Secret %q deleted.", args[0])}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q deleted.\n", args[0])
		})
		return nil
	},
}

func init() {
	vaultDeleteCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	registerCommand(vaultDeleteCmd)
}
