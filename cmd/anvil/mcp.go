package main

import (
	"github.com/inovacc/anvil/internal/mcpserver"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio transport)",
	Long:  "Start an MCP server exposing vault operations as tools. Communicates via stdin/stdout JSON-RPC.",
	RunE: func(cmd *cobra.Command, args []string) error {
		srv, err := mcpserver.New(nil)
		if err != nil {
			return err
		}
		return srv.Run(cmd.Context())
	},
}

func init() {
	registerCommand(mcpCmd)
}
