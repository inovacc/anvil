package cmd

import (
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show vault status",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, err := vault.GetStatus(nil)
		if err != nil {
			return err
		}

		outputResult(cmd, status, func() {
			if !status.Initialized {
				fmt.Println("Vault is not initialized.")
				fmt.Printf("Database path: %s\n", status.DBPath)
				fmt.Println("Run 'profile vault init' to initialize.")
				return
			}

			fmt.Println("Vault Status")
			fmt.Println("  Initialized:  yes")
			fmt.Printf("  Database:     %s\n", status.DBPath)
			fmt.Printf("  Profiles:     %d\n", status.ProfileCount)
			fmt.Printf("  Secrets:      %d\n", status.SecretCount)
			fmt.Printf("  Key version:  %d\n", status.KeyVersion)
			if !status.CreatedAt.IsZero() {
				fmt.Printf("  Created:      %s\n", status.CreatedAt.Format("2006-01-02 15:04:05"))
			}
		})
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultStatusCmd)
}
