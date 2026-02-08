package cmd

import (
	"fmt"
	"time"

	"github.com/inovacc/profile/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultEnvReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release secrets for a time-limited period",
	Long:  "Authenticates with password and releases vault secrets as environment variables for the specified duration.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		password, err := readPassword("Password: ")
		if err != nil {
			return err
		}

		ttl, _ := cmd.Flags().GetDuration("ttl")
		profileName, _ := cmd.Flags().GetString("profile")

		state, err := v.EnvRelease(password, &vault.EnvReleaseOptions{
			ProfileName: profileName,
			TTL:         ttl,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Secrets released for profile %q. Expires in %s\n", state.ProfileName, state.Remaining.Truncate(time.Second))
		return nil
	},
}

func init() {
	vaultEnvReleaseCmd.Flags().DurationP("ttl", "t", 30*time.Minute, "Time-to-live for the release (min 1m, max 24h)")
	vaultEnvReleaseCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultEnvCmd.AddCommand(vaultEnvReleaseCmd)
}
