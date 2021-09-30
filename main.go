package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/samjwillis97/GoTDMS/cmd"
	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	log "github.com/sirupsen/logrus"
)

const versionString string = "0.0.1.0"

func main() {
	cmd.Execute()
}

func old_main() {
	var debug bool
	var help bool
	var json bool
	var verbose bool
	var version bool
	var timed bool

	//TODO: Implement JSON Outputs

	flag.BoolVar(&json, "json", false, "Output in JSON Format")
	flag.BoolVar(&debug, "d", false, "Output Debug Log")
	flag.BoolVar(&debug, "debug", false, "Output Debug Log")
	flag.BoolVar(&help, "h", false, "Help")
	flag.BoolVar(&help, "help", false, "Help")
	flag.BoolVar(&verbose, "v", false, "Verbose")
	flag.BoolVar(&verbose, "verbose", false, "Verbose")
	flag.BoolVar(&version, "version", false, "Version")
	flag.BoolVar(&timed, "t", false, "Time")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
	} else if version {
		printVersion()
		os.Exit(0)
	}

	var startTime time.Time

	if timed {
		startTime = time.Now()
	}

	args := flag.Args()

	initLogging(debug)

	if len(args) < 1 {
		printHelp()
		os.Exit(1)
	}

	switch args[0] {
	case "list":
		switch args[1] {
		case "groups":
			if len(args) < 3 {
				fmt.Println("File Path Missing")
				log.Fatal("File path missing")
			} else {
				filePath := args[2]
				log.Debugln("Opening: ", filePath)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					fmt.Println("File does not exist")
					log.Fatal("File does not exist")
				} else {
					file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
					if err != nil {
						fmt.Println("Error Opening TDMS File")
						log.Fatal("Error opening TDMS File")
					}
					tdms.DisplayGroups(file)
					file.Close()
				}
			}
		case "channels":
			log.Debugln(len(os.Args))
			if len(args) < 4 {
				fmt.Println("Not Enough Arguments Supplied, expected 2")
				log.Fatal("Group/File Missing")
			} else {
				groupName := args[2]
				filePath := args[3]
				log.Debugln("Opening: ", filePath)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					fmt.Println("File does not exist")
					log.Fatal("File does not exist")
				} else {
					file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
					if err != nil {
						fmt.Println("Error Opening TDMS File")
						log.Fatal("Error opening TDMS File")
					}
					tdms.DisplayGroupChannels(file, groupName)
					file.Close()
				}
			}
		case "properties":
			// Have a switch for Channel or Group properties
			if len(args) < 5 {
				fmt.Println("Not Enough Arguments Supplied, expected 3")
				log.Fatal("Group/Channel/File Missing")
			} else {
				groupName := args[2]
				channelName := args[3]
				filePath := args[4]
				log.Debugln("Opening: ", filePath)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					fmt.Println("File does not exist")
					log.Fatal("File does not exist")
				} else {
					file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
					if err != nil {
						fmt.Println("Error Opening TDMS File")
						log.Fatal("Error opening TDMS File")
					}
					tdms.DisplayChannelProperties(file, groupName, channelName)
					file.Close()
				}
			}
		default:

			if len(args) < 2 {
				fmt.Println("Unkown Object to List")
				fmt.Println()
				printListHelp()
			} else if _, err := os.Stat(args[1]); os.IsNotExist(err) {
				fmt.Println("Unkown Object to List")
				fmt.Println()
				printListHelp()
			} else {
				file, err := os.OpenFile(args[1], os.O_RDONLY, 0666)
				if err != nil {
					fmt.Println("Error Opening TDMS File")
					log.Fatal("Error opening TDMS File")
				}
				tdms.DisplayFile(file, verbose)
				file.Close()
			}
		}
	case "read":
		switch args[1] {
		case "channel":
			if len(args) < 5 {
				fmt.Println("Not Enough Arguments Supplied, expected 3")
				log.Fatal("Group/Channel/File path missing")
			} else {
				groupName := args[2]
				channelName := args[3]
				filePath := args[4]
				log.Debugln("Opening: ", filePath)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					fmt.Println("File does not exist")
					log.Fatal("File does not exist")
				} else {
					file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
					if err != nil {
						fmt.Println("Error Opening TDMS File")
						log.Fatal("Error opening TDMS File")
					}
					tdms.DisplayChannelData(file, groupName, channelName)
					file.Close()
				}
			}
		}
	default:
		fmt.Println("Unkown Command")
		printHelp()
	}
	if timed {
		elapsed := time.Since(startTime)
		fmt.Println()
		fmt.Println("Execution Time: ", elapsed)
	}
}

func printVersion() {
	fmt.Println("Version: ", versionString)
}

func printHelp() {
	fmt.Println("NAME:")
	fmt.Println("  GoTDMS - Command-line TDMS Reader written in Go")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  GoTDMS [global options] command subcommand [options] [arguments...]")
	fmt.Println()
	fmt.Println("VERSION:")
	fmt.Print("  ")
	fmt.Println(versionString)
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("    help")
	fmt.Println("    list")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --debug, -d")
	fmt.Println("  --help, -h")
	fmt.Println("  --json")
	fmt.Println("  --verbose, -v")
	fmt.Println("  --version")
}

func printListHelp() {
	fmt.Println("USAGE:")
	fmt.Println("  GoTDMS [global options] list subcommand [options] [arguments...]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  groups")
	fmt.Println("  channels")
	fmt.Println("  properties")
}

func printReadHelp() {
	fmt.Println("USAGE:")
	fmt.Println("  GoTDMS [global options] read subcommand [options] [arguments...]")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  channel")
}

func initLogging(debug bool) {
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

	if debug {
		log.SetLevel(log.DebugLevel)
	}
	// log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}
