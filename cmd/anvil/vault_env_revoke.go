package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultEnvRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke the active env release",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		if err := v.EnvRevoke(); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{"Release revoked."}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Release revoked.")
		})

		return nil
	},
}

func init() {
	vaultEnvCmd.AddCommand(vaultEnvRevokeCmd)
}
