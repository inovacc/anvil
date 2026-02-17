package main

import (
	"fmt"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultPluginHookAddCmd = &cobra.Command{
	Use:   "hook-add <event> <command> [args...]",
	Short: "Add a hook for a vault event",
	Long: `Add a hook that runs when a vault event occurs.

Events: pre-set, post-set, pre-get, post-get, pre-delete, post-delete, post-rotate, post-init

Example:
  anvil vault plugin hook-add post-set /usr/local/bin/notify --channel ops`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		event := vault.HookEvent(args[0])
		command := args[1]
		var hookArgs []string
		if len(args) > 2 {
			hookArgs = args[2:]
		}

		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return err
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")
		pm := vault.NewPluginManager(configPath)
		pm.AddHook(event, command, hookArgs)

		if err := pm.SaveConfig(); err != nil {
			return err
		}

		outputResult(cmd, map[string]string{"status": "ok", "event": string(event), "command": command}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Hook added: [%s] %s\n", event, command)
		})

		return nil
	},
}

func init() {
	vaultPluginCmd.AddCommand(vaultPluginHookAddCmd)
}
