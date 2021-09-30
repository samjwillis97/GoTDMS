package cmd

import (
	"os"

	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listPropertiesCmd)
}

var listPropertiesCmd = &cobra.Command{
	Use:   "list-properties [file] [group] [channel]",
	Short: "List the properties of the given channel of a TDMS File",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		groupName := args[1]
		channelName := args[2]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		tdms.DisplayChannelProperties(file, groupName, channelName)
		file.Close()
		return nil
	},
}
