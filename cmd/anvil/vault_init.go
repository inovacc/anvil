package cmd

import (
	"errors"
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the encrypted vault",
	Long:  "Creates a new vault with a machine-bound master key. Only needs to be run once.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vault.Init(nil); err != nil {
			if errors.Is(err, vault.ErrAlreadyInitialized) {
				outputResult(cmd, struct {
					Initialized bool   `json:"initialized"`
					DBPath      string `json:"db_path"`
					Message     string `json:"message"`
				}{true, vault.DefaultDBPath(), "Vault is already initialized."}, func() {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Vault is already initialized.")
				})
				return nil
			}
			return err
		}

		outputResult(cmd, struct {
			Initialized bool   `json:"initialized"`
			DBPath      string `json:"db_path"`
			Message     string `json:"message"`
		}{true, vault.DefaultDBPath(), "Vault initialized successfully."}, func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, "Vault initialized successfully.")
			_, _ = fmt.Fprintf(w, "Database: %s\n", vault.DefaultDBPath())
		})
		return nil
	},
}

func init() {
	registerCommand(vaultInitCmd)
}
