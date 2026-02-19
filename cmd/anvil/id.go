package main

import (
	"fmt"

	"github.com/inovacc/anvil/pkg/vault"
	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Show the machine-bound installation ID",
	Long:  "Display a deterministic identifier derived from the machine identity and vault sealed key.",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := vault.Open(nil)
		if err != nil {
			handleError(cmd, err)
			return err
		}
		defer func() { _ = v.Close() }()

		id, err := v.InstallationID()
		if err != nil {
			handleError(cmd, err)
			return err
		}

		outputResult(cmd, struct {
			InstallationID string `json:"installation_id"`
		}{id}, func() {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
		})

		return nil
	},
}

func init() {
	registerCommand(idCmd)
}
