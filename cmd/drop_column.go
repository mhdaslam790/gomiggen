package cmd

import (
	"github.com/mhdaslam790/gomiggen/internal"
	"github.com/spf13/cobra"
)

var dropColumnCmd = &cobra.Command{
	Use:   "drop-column [ModelName] [ColumnName]",
	Short: "Drop a column from a given model",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		internal.HandleAction([]string{"drop-column", args[0], args[1]})
	},
}
