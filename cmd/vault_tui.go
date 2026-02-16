package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/inovacc/anvil/internal/tui"
	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultTUICmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI for vault management",
	Long:  "Launch an interactive terminal interface for browsing profiles, managing secrets, and viewing vault status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		m := tui.NewModel(v)
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(vaultTUICmd)
}
