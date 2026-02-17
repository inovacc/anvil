package cmd

import "github.com/spf13/cobra"

// registerCommand adds cmd to rootCmd (new flat CLI structure).
// Backward compatibility for `anvil vault ...` is handled by vaultCmd delegation.
func registerCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
