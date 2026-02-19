package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var keyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			return err
		}

		defer func() { _ = v.Close() }()

		keys, err := v.ListKeys()
		if err != nil {
			return err
		}

		outputResult(cmd, keys, func() {
			w := cmd.OutOrStdout()
			if len(keys) == 0 {
				_, _ = fmt.Fprintln(w, "No keys found.")
				return
			}

			tw := tableWriter(w)
			_, _ = fmt.Fprintln(tw, "NAME\tALGORITHM\tFINGERPRINT\tDESCRIPTION\tCREATED")

			for _, k := range keys {
				_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					k.Name,
					k.Algorithm,
					k.Fingerprint,
					k.Description,
					k.CreatedAt.Format("2006-01-02 15:04:05"),
				)
			}

			_ = tw.Flush()
		})

		return nil
	},
}

func init() {
	keyCmd.AddCommand(keyListCmd)
}
