package cmd

import (
	"fmt"
	"strconv"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultHistoryCmd = &cobra.Command{
	Use:   "history <key>",
	Short: "Show version history for a secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		versions, err := v.SecretHistory(args[0], profileName)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			outputResult(cmd, []vault.SecretVersion{}, func() {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "No version history for %q.\n", args[0])
			})
			return nil
		}

		outputResult(cmd, versions, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Version history for %q (%d versions):\n\n", args[0], len(versions))
			tw := tableWriter(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "VERSION\tCREATED")
			for _, ver := range versions {
				_, _ = fmt.Fprintf(tw, "v%d\t%s\n", ver.Version, ver.CreatedAt.Format("2006-01-02 15:04:05"))
			}
			_ = tw.Flush()
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nUse 'anvil vault rollback %s <version>' to restore.\n", args[0])
		})
		return nil
	},
}

var vaultRollbackCmd = &cobra.Command{
	Use:   "rollback <key> <version>",
	Short: "Rollback a secret to a previous version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")

		version, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}

		if err := v.SecretRollback(args[0], profileName, version); err != nil {
			return err
		}

		outputResult(cmd, struct {
			Key     string `json:"key"`
			Version int64  `json:"version"`
			Message string `json:"message"`
		}{args[0], version, fmt.Sprintf("Secret %q rolled back to version %d.", args[0], version)}, func() {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Secret %q rolled back to version %d.\n", args[0], version)
		})
		return nil
	},
}

func init() {
	vaultHistoryCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	vaultRollbackCmd.Flags().StringP("profile", "p", "", "Target profile (default: active profile)")
	registerCommand(vaultHistoryCmd)
	registerCommand(vaultRollbackCmd)
}
