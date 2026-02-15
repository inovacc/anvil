package cmd

import "github.com/spf13/cobra"

var vaultPluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage vault plugins (hooks and providers)",
	Long:  "Add, remove, and list hooks and secret providers for the vault plugin system.",
}

func init() {
	vaultCmd.AddCommand(vaultPluginCmd)
}
