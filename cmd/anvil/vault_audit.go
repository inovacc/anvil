package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var vaultAuditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Show audit log entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}
		defer func() { _ = v.Close() }()

		profileName, _ := cmd.Flags().GetString("profile")
		limit, _ := cmd.Flags().GetInt64("limit")

		var entries []vault.AuditEntry

		if profileName != "" {
			entries, err = v.AuditLogByProfile(profileName, limit)
		} else {
			entries, err = v.AuditLog(limit)
		}

		if err != nil {
			return err
		}

		outputResult(cmd, entries, func() {
			if len(entries) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No audit log entries.")
				return
			}

			tw := tableWriter(cmd.OutOrStdout())
			_, _ = fmt.Fprintln(tw, "TIMESTAMP\tACTION\tPROFILE\tKEY\tDETAIL")
			for _, e := range entries {
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					e.CreatedAt.Format("2006-01-02 15:04:05"),
					e.Action,
					e.ProfileName,
					e.SecretKey,
					e.Detail,
				)
			}
			_ = tw.Flush()
		})

		return nil
	},
}

func init() {
	vaultAuditCmd.Flags().StringP("profile", "p", "", "Filter by profile name")
	vaultAuditCmd.Flags().Int64P("limit", "l", 50, "Maximum entries to show")
	registerCommand(vaultAuditCmd)
}
