package cmd

import (
	"os"

	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list [file]",
	Short: "List the full TDMS file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		tdms.DisplayFile(file, Verbose)
		file.Close()
		return nil
	},
}
