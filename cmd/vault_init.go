package cmd

import (
	"errors"
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the encrypted vault",
	Long:  "Creates a new vault with a machine-bound master key. Only needs to be run once.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := vault.Init(nil); err != nil {
			if errors.Is(err, vault.ErrAlreadyInitialized) {
				fmt.Println("Vault is already initialized.")
				return nil
			}
			return err
		}

		fmt.Println("Vault initialized successfully.")
		fmt.Printf("Database: %s\n", vault.DefaultDBPath())
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultInitCmd)
}
