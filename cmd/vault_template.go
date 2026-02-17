package cmd

import "github.com/spf13/cobra"

var vaultTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage secret templates",
	Long:  "Create, list, show, delete, and apply secret templates for quick secret provisioning.",
}

func init() {
	registerCommand(vaultTemplateCmd)
}
