package cmd

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultSealCmd = &cobra.Command{
	Use:   "seal",
	Short: "Temporarily lock the vault",
	Long:  "Seal the vault to prevent all read and write operations until unsealed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		if err := v.Seal(); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{"Vault sealed."}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vault sealed. All operations are locked until unsealed.")
		})
		return nil
	},
}

var vaultUnsealCmd = &cobra.Command{
	Use:   "unseal",
	Short: "Unlock a sealed vault",
	Long:  "Unseal the vault to restore read and write operations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vault.UnsealVault(nil); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Message string `json:"message"`
		}{"Vault unsealed."}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vault unsealed. Operations restored.")
		})
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultSealCmd)
	vaultCmd.AddCommand(vaultUnsealCmd)
}
