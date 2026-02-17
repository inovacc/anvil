package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultPluginProviderAddCmd = &cobra.Command{
	Use:   "provider-add <name> <command> <prefix>",
	Short: "Add a secret provider",
	Long: `Register an external secret provider.

Example:
  anvil vault plugin provider-add aws /usr/local/bin/aws-provider aws/`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, command, prefix := args[0], args[1], args[2]

		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return err
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")
		pm := vault.NewPluginManager(configPath)
		pm.AddProvider(name, command, prefix)

		if err := pm.SaveConfig(); err != nil {
			return err
		}

		outputResult(cmd, map[string]string{"status": "ok", "name": name}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Provider added: %s (prefix: %s)\n", name, prefix)
		})

		return nil
	},
}

func init() {
	vaultPluginCmd.AddCommand(vaultPluginProviderAddCmd)
}
