package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migration",
	Short: "A CLI to generate GORM migrations automatically",
	Long:  "A Go CLI tool for generating database migrations using GORM, with built-in model support and Cobra-based commands.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	rootCmd.AddCommand(addColumnCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(addIndexCmd)
	rootCmd.AddCommand(dropColumnCmd)
	rootCmd.AddCommand(dropIndexCmd)
}
