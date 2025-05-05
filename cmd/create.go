package cmd

import (
	"github.com/mhdaslam790/gomiggen/internal"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:     "create [ModelName]",
	Short:   "Create a new model & migration",
	Aliases: []string{"c"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		internal.HandleAction([]string{"create", args[0]})
	},
}
