package cmd

import (
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Manage encrypted secrets",
	Long:  "Machine-bound encrypted vault for storing sensitive key-value secrets organized by profiles.",
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}
