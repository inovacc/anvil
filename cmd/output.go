package cmd

import (
	"encoding/json"
	"fmt"

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
		fmt.Println(string(data))
		return
	}
	textFn()
}
