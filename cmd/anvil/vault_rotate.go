package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultRotateKeyCmd = &cobra.Command{
	Use:   "rotate-key",
	Short: "Rotate the vault master key",
	Long:  "Generates a new master key, re-encrypts all secrets, and re-seals the key. Requires password verification. Any active env release session is revoked.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			var err error
			password, err = readPassword("Password: ")
			if err != nil {
				return err
			}
		}

		if err := v.RotateKey(password); err != nil {
			return err
		}

		status, err := v.Status()
		if err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message    string `json:"message"`
			KeyVersion int64  `json:"key_version"`
			SealMethod string `json:"seal_method"`
		}{
			Message:    fmt.Sprintf("Master key rotated successfully (version %d)", status.KeyVersion),
			KeyVersion: status.KeyVersion,
			SealMethod: status.SealMethod,
		}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Master key rotated successfully (version %d)\n", status.KeyVersion)
		})

		return nil
	},
}

func init() {
	vaultRotateKeyCmd.Flags().String("password", "", "Password (non-interactive mode)")
	registerCommand(vaultRotateKeyCmd)
}
