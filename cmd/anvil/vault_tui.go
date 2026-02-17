package main

import (
	"github.com/spf13/cobra"
)

var vaultTUICmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive terminal UI for vault management",
	Long:  "Launch an interactive terminal interface for browsing profiles, managing secrets, and viewing vault status.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchTUI()
	},
}

func init() {
	registerCommand(vaultTUICmd)
}
