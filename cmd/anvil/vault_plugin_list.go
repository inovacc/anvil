package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultPluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured hooks and providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, err := vault.ResolveDBPath(nil)
		if err != nil {
			return err
		}

		configPath := filepath.Join(filepath.Dir(dbPath), "plugins.json")
		pm := vault.NewPluginManager(configPath)
		cfg := pm.Config()

		type output struct {
			Hooks     []vault.HookConfig     `json:"hooks"`
			Providers []vault.ProviderConfig  `json:"providers"`
		}

		out := output{
			Hooks:     cfg.Hooks,
			Providers: cfg.Providers,
		}
		if out.Hooks == nil {
			out.Hooks = []vault.HookConfig{}
		}
		if out.Providers == nil {
			out.Providers = []vault.ProviderConfig{}
		}

		outputResult(cmd, out, func() {
			w := cmd.OutOrStdout()

			_, _ = fmt.Fprintln(w, "Hooks:")
			if len(cfg.Hooks) == 0 {
				_, _ = fmt.Fprintln(w, "  (none)")
			}
			for _, h := range cfg.Hooks {
				_, _ = fmt.Fprintf(w, "  [%s] %s %v\n", h.Event, h.Command, h.Args)
			}

			_, _ = fmt.Fprintln(w, "\nProviders:")
			if len(cfg.Providers) == 0 {
				_, _ = fmt.Fprintln(w, "  (none)")
			}
			for _, p := range cfg.Providers {
				_, _ = fmt.Fprintf(w, "  %s (prefix: %s) → %s\n", p.Name, p.Prefix, p.Command)
			}
		})

		return nil
	},
}

func init() {
	vaultPluginCmd.AddCommand(vaultPluginListCmd)
}
