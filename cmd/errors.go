package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

func handleError(cmd *cobra.Command, err error) {
	useJSON, _ := cmd.Flags().GetBool("json")
	w := cmd.ErrOrStderr()

	if ue, ok := vault.IsUserError(err); ok {
		if useJSON {
			data := struct {
				Error string `json:"error"`
				Hint  string `json:"hint,omitempty"`
			}{Error: ue.Message, Hint: ue.Hint}

			out, _ := json.MarshalIndent(data, "", "  ")
			_, _ = fmt.Fprintln(w, string(out))

			return
		}

		_, _ = fmt.Fprintf(w, "Error: %s\n", ue.Message)

		if ue.Hint != "" {
			_, _ = fmt.Fprintf(w, "Hint:  %s\n", ue.Hint)
		}

		return
	}

	if useJSON {
		data := struct {
			Error string `json:"error"`
		}{Error: err.Error()}

		out, _ := json.MarshalIndent(data, "", "  ")
		_, _ = fmt.Fprintln(w, string(out))

		return
	}

	_, _ = fmt.Fprintf(w, "Error: %s\n", err)
}
