package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
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

		fmt.Println("Release revoked.")
		return nil
	},
}

func init() {
	vaultEnvCmd.AddCommand(vaultEnvRevokeCmd)
}
