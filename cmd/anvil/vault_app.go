package main

import (
	"github.com/spf13/cobra"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage isolated per-app vaults",
}

func init() {
	registerCommand(appCmd)
}
