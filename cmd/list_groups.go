package cmd

import (
	"os"

	"github.com/samjwillis97/GoTDMS/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listGroupsCmd)
}

var listGroupsCmd = &cobra.Command{
	Use:   "list-groups [file]",
	Short: "List the groups in the given TDMS File",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verify Arg one is file
		filePath := args[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		cli.DisplayGroups(file)
		file.Close()
		return nil
	},
}
