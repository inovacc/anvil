package main

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// outputResult checks the --json flag and prints either JSON or human-readable text.
func outputResult(cmd *cobra.Command, jsonData any, textFn func()) {
	useJSON, _ := cmd.Flags().GetBool("json")
	if useJSON {
		data, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error marshaling JSON: %v\n", err)
			return
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(data))

		return
	}

	textFn()
}

// tableWriter creates a tabwriter with consistent settings for CLI table output.
func tableWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
}
