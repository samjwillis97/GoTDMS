package cmd

import (
	"os"

	"github.com/samjwillis97/GoTDMS/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listChannelsCmd)
}

var listChannelsCmd = &cobra.Command{
	Use:   "list-channels [file] [group]",
	Short: "List the channels in the given group of a TDMS File",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Verify Arg one is file
		filePath := args[0]
		groupName := args[1]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		cli.DisplayGroupChannels(file, groupName)
		file.Close()
		return nil
	},
}
