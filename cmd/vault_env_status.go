package cmd

import (
	"fmt"
	"time"

	"github.com/inovacc/profile/pkg/vault"
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
			if !state.Active {
				fmt.Println("Status:  inactive")
				return
			}

			fmt.Printf("Status:    active\n")
			fmt.Printf("Profile:   %s\n", state.ProfileName)
			fmt.Printf("Session:   %s\n", state.SessionID)
			fmt.Printf("Expires:   %s\n", state.ExpiresAt.Format(time.RFC3339))
			fmt.Printf("Remaining: %s\n", state.Remaining.Truncate(time.Second))
		})
		return nil
	},
}

func init() {
	vaultEnvCmd.AddCommand(vaultEnvStatusCmd)
}
