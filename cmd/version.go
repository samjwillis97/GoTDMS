package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GoTDMS",
	Long:  "All software has versions. This is GoTDMS'",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoTDMS Reader v0.0.1")
	},
}
