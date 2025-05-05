package cmd

import (
	"github.com/mhdaslam790/gomiggen/internal"

	"github.com/spf13/cobra"
)

var dropIndexCmd = &cobra.Command{
	Use:   "drop-index [ModelName] [ColumnName]",
	Short: "Drop an index from a column in a given model",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		internal.HandleAction([]string{"drop-index", args[0], args[1]})
	},
}
