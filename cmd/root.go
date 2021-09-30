package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hugo",
	Short: "Hugo is a very fast static site generator",
	Long:  "Longer Description Please",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
}
