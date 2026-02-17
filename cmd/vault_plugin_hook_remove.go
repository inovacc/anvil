package main

import (
	"fmt"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultPluginHookRemoveCmd = &cobra.Command{
	Use:   "hook-remove <event> <command>",
	Short: "Remove a hook for a vault event",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		event := vault.HookEvent(args[0])
		command := args[1]

		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return err
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")
		pm := vault.NewPluginManager(configPath)
		if err := pm.RemoveHook(event, command); err != nil {
			return err
		}

		if err := pm.SaveConfig(); err != nil {
			return err
		}

		outputResult(cmd, map[string]string{"status": "ok", "event": string(event), "command": command}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Hook removed: [%s] %s\n", event, command)
		})

		return nil
	},
}

func init() {
	vaultPluginCmd.AddCommand(vaultPluginHookRemoveCmd)
}
