package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a decrypted secret value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		value, err := v.Get(args[0], profileName)
		if err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{args[0], value}, func() {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), value)
		})
		return nil
	},
}

func init() {
	vaultGetCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	registerCommand(vaultGetCmd)
}
