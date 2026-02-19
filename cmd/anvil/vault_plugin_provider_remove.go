package main

import (
	"fmt"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultPluginProviderRemoveCmd = &cobra.Command{
	Use:   "provider-remove <name>",
	Short: "Remove a secret provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return err
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")

		pm := vault.NewPluginManager(configPath)
		if err := pm.RemoveProvider(name); err != nil {
			return err
		}

		if err := pm.SaveConfig(); err != nil {
			return err
		}

		outputResult(cmd, map[string]string{"status": "ok", "name": name}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Provider removed: %s\n", name)
		})

		return nil
	},
}

func init() {
	vaultPluginCmd.AddCommand(vaultPluginProviderRemoveCmd)
}
