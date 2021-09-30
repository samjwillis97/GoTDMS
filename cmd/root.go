package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	Verbose bool
	Json    bool
	Debug   bool
	Timed   bool

	StartTime time.Time

	rootCmd = &cobra.Command{
		Use:     "gotdms",
		Short:   "GoTDMS is a Command Line NI TDMS File Reader",
		PostRun: postRunFunc,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initFunction)

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&Json, "json", "j", false, "json formatted output")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().BoolVarP(&Timed, "timed", "t", false, "use for timing")
}

func initFunction() {
	initLogging()
	if Timed {
		StartTime = time.Now()
	}
}

func initLogging() {
	// If the file doesnt exit create it, or append to the file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error return from os.OpenFile: ", err)
	}

	// Creates a MultiWrite, duplicates its write sto all the provided writers
	// In this instance, the file and stdout
	// mw := io.MultiWriter(os.Stdout, file)
	mw := io.MultiWriter(file)

	log.SetOutput(mw)

	if Debug {
		log.SetLevel(log.DebugLevel)
	}
	// log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

func postRunFunc(cmd *cobra.Command, args []string) {
	//TODO: Fixed Time
	if Timed {
		elapsed := time.Since(StartTime)
		fmt.Println()
		fmt.Println("Execution Time: ", elapsed)
	}
}
