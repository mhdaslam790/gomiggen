package cmd

import (
	"github.com/mhdaslam790/gomiggen/internal"
	"github.com/spf13/cobra"
)

var addIndexCmd = &cobra.Command{
	Use:   "add-index [ModelName] [ColumnName]",
	Short: "Add an index on a column for a given model",
	Args:  cobra.ExactArgs(2),
	
	Run: func(cmd *cobra.Command, args []string) {
		internal.HandleAction([]string{"add-index", args[0], args[1]})
	},
}

