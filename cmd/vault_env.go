package cmd

import (
	"github.com/spf13/cobra"
)

var vaultEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage password-gated environment variable release",
	Long:  "Control time-limited access to vault secrets as environment variables. Requires a password gate.",
}

func init() {
	vaultCmd.AddCommand(vaultEnvCmd)
}
