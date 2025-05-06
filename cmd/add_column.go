package cmd

import (
	"github.com/mhdaslam790/gomiggen/internal"
	"github.com/spf13/cobra"
)

var addColumnCmd = &cobra.Command{
	Use:   "add-column [ModelName] [column:type]",
	Short: "Add a column to a model's migration",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		internal.HandleAction([]string{"add-column", args[0], args[1]})
	},
}
