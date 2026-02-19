package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type cmdTreeEntry struct {
	Name        string `json:"name"`
	FullPath    string `json:"full_path"`
	Description string `json:"short_description"`
}

var cmdTreeCmd = &cobra.Command{
	Use:   "cmdtree",
	Short: "Display command tree",
	Long:  "Shows a visual tree of all available commands.",
	RunE: func(cmd *cobra.Command, args []string) error {
		outputResult(cmd, flattenCommands(rootCmd, ""), func() {
			w := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(w, rootCmd.Name())
			printTree(w, rootCmd, "")
		})

		return nil
	},
}

func printTree(w io.Writer, cmd *cobra.Command, prefix string) {
	children := visibleSubcommands(cmd)
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├── "
		childPrefix := "│   "

		if isLast {
			connector = "└── "
			childPrefix = "    "
		}

		_, _ = fmt.Fprintf(w, "%s%s%s\n", prefix, connector, child.Name())
		printTree(w, child, prefix+childPrefix)
	}
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	var visible []*cobra.Command

	for _, c := range cmd.Commands() {
		if !c.Hidden && c.Name() != "help" {
			visible = append(visible, c)
		}
	}

	return visible
}

func flattenCommands(cmd *cobra.Command, parentPath string) []cmdTreeEntry {
	fullPath := cmd.Name()
	if parentPath != "" {
		fullPath = parentPath + " " + cmd.Name()
	}

	children := visibleSubcommands(cmd)
	entries := make([]cmdTreeEntry, 1, 1+len(children))
	entries[0] = cmdTreeEntry{
		Name:        cmd.Name(),
		FullPath:    strings.TrimSpace(fullPath),
		Description: cmd.Short,
	}

	for _, child := range children {
		entries = append(entries, flattenCommands(child, fullPath)...)
	}

	return entries
}

func init() {
	rootCmd.AddCommand(cmdTreeCmd)
}
