package main

import (
	"fmt"
	"os"

	"github.com/inovacc/anvil/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// readPassword reads a password from the terminal without echoing.
func readPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)

	fd := int(os.Stdin.Fd())

	password, err := term.ReadPassword(fd)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	_, _ = fmt.Fprintln(os.Stderr)

	return string(password), nil
}

// getOutputOpts reads the global --json and --table flags from the command
// and returns an output.Options. During migration, commands that still
// declare local --json flags will shadow the persistent one — Cobra resolves
// local flags first, so there is no breakage.
func getOutputOpts(cmd *cobra.Command) output.Options {
	j, _ := cmd.Flags().GetBool("json")
	tbl, _ := cmd.Flags().GetBool("table")

	return output.Options{JSON: j, Table: tbl}
}
