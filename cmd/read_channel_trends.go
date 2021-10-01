package cmd

import (
	"os"

	"github.com/samjwillis97/GoTDMS/pkg/cli"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(readChannelTrendsCmd)
}

var readChannelTrendsCmd = &cobra.Command{
	Use:   "read-channel-trends [file] [group] [channel]",
	Short: "Outputs trend information from the chosen channel",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		groupName := args[1]
		chanName := args[2]
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return err
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		cli.DisplayChannelData(file, groupName, chanName)
		file.Close()
		return nil
	},
}
