package main

import (
	"github.com/spf13/cobra"
)

var vaultCmd = &cobra.Command{
	Use:                "vault",
	Short:              "Manage encrypted secrets",
	Long:               "Machine-bound encrypted vault for storing sensitive key-value secrets organized by profiles.",
	Hidden:             true,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		rootCmd.SetArgs(args)
		return rootCmd.Execute()
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)
}
