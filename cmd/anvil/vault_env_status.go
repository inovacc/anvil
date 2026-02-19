package main

import (
	"fmt"
	"time"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultEnvStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current env release status",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		state, err := v.EnvStatus()
		if err != nil {
			return err
		}

		outputResult(cmd, struct {
			Active    bool   `json:"active"`
			Profile   string `json:"profile,omitempty"`
			SessionID string `json:"session_id,omitempty"`
			ExpiresAt string `json:"expires_at,omitempty"`
			Remaining string `json:"remaining,omitempty"`
		}{
			Active:    state.Active,
			Profile:   state.ProfileName,
			SessionID: state.SessionID,
			ExpiresAt: func() string {
				if state.Active {
					return state.ExpiresAt.Format(time.RFC3339)
				}

				return ""
			}(),
			Remaining: func() string {
				if state.Active {
					return state.Remaining.Truncate(time.Second).String()
				}

				return ""
			}(),
		}, func() {
			w := cmd.OutOrStdout()
			if !state.Active {
				_, _ = fmt.Fprintln(w, "Status:  inactive")
				return
			}

			_, _ = fmt.Fprintf(w, "Status:    active\n")
			_, _ = fmt.Fprintf(w, "Profile:   %s\n", state.ProfileName)
			_, _ = fmt.Fprintf(w, "Session:   %s\n", state.SessionID)
			_, _ = fmt.Fprintf(w, "Expires:   %s\n", state.ExpiresAt.Format(time.RFC3339))
			_, _ = fmt.Fprintf(w, "Remaining: %s\n", state.Remaining.Truncate(time.Second))
		})

		return nil
	},
}

func init() {
	vaultEnvCmd.AddCommand(vaultEnvStatusCmd)
}
