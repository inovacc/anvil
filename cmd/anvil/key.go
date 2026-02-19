package main

import (
	"github.com/spf13/cobra"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Manage asymmetric keys",
	Long:  "Generate, list, delete, export, and import Ed25519 and ECDSA P-256 key pairs.",
}

func init() {
	registerCommand(keyCmd)
}
