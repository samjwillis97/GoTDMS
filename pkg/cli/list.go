package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	log "github.com/sirupsen/logrus"
)

func DisplayFile(file *os.File, verbose bool) {
	// Get All Segments, Find all Non Duplicates
	// Get Each Group, Each Channel and All Properties
	segments, _ := tdms.ReadAllSegments(file)

	paths := tdms.ReadAllUniqueTDMSObjects(segments)

	groups := tdms.GetGroupsFromPathArray(paths)

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	for groupIter, group := range groups {
		// GROUPS
		formattedG := strings.Replace(group, "/", "", -1)
		formattedG = strings.Replace(formattedG, "'", "", -1)

		if groupIter == len(groups)-1 {
			fmt.Fprintf(writer, "└── %s\n", formattedG)
		} else {
			fmt.Fprintf(writer, "├── %s\n", formattedG)
		}

		channels := tdms.GetChannelsFromPathArray(paths, formattedG)

		for chanIter, channel := range channels {
			// CHANNELS
			formattedC := strings.Replace(channel, "/", "", -1)
			formattedC = strings.Replace(formattedC, "'", "", -1)

			var chFormatString strings.Builder

			if groupIter != len(groups)-1 {
				chFormatString.WriteString("|")
			}
			chFormatString.WriteString("\t")
			if chanIter == len(channels)-1 {
				chFormatString.WriteString("└──")
			} else {
				chFormatString.WriteString("├──")
			}
			chFormatString.WriteString(" %s\n")

			fmt.Fprintf(writer, chFormatString.String(), formattedC)

			var properties tdms.Properties
			for _, val := range segments {
				//PROPERTIES
				for path, propMap := range val.PropMap {
					if path == ("/'" + formattedG + "'/'" + formattedC + "'") {
						for _, propValue := range propMap {
							properties = append(properties, propValue)
						}
					}
				}
			}

			sort.Sort(properties)
			if verbose {
				for propIter, val := range properties {

					var propFormatString strings.Builder

					if groupIter != len(groups)-1 {
						propFormatString.WriteString("|")
					}
					propFormatString.WriteString("\t")
					if chanIter != len(channels)-1 {
						propFormatString.WriteString("|")
					}
					propFormatString.WriteString("\t")
					if propIter == len(properties)-1 {
						propFormatString.WriteString("└──")
					} else {
						propFormatString.WriteString("├──")
					}
					propFormatString.WriteString(" %s\t%s\n")

					fmt.Fprintf(writer, propFormatString.String(), val.Name, val.StringValue)
				}
			}
		}
	}
	writer.Flush()
}

func DisplayGroups(file *os.File) {
	segments, _ := tdms.ReadAllSegments(file)

	paths := tdms.ReadAllUniqueTDMSObjects(segments)

	groups := tdms.GetGroupsFromPathArray(paths)

	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		fmt.Println(formatted)
	}
}

func DisplayGroupChannels(file *os.File, groupName string) {
	segments, _ := tdms.ReadAllSegments(file)

	paths := tdms.ReadAllUniqueTDMSObjects(segments)

	groups := tdms.GetGroupsFromPathArray(paths)

	groupPresent := false
	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		if formatted == groupName {
			groupPresent = true
			log.Debugf("Found matching Group")
		}
	}

	var channels []string

	if groupPresent {
		channels = tdms.GetChannelsFromPathArray(paths, groupName)
	} else {
		fmt.Println("File does not contain group named: ", groupName)
		log.Fatal("Invalid Group Name")
	}

	for _, channel := range channels {
		formatted := strings.Replace(channel, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		fmt.Println(formatted)
	}
}

func DisplayChannelProperties(file *os.File, groupName string, channelName string) {
	segments, _ := tdms.ReadAllSegments(file)

	paths := tdms.ReadAllUniqueTDMSObjects(segments)

	groups := tdms.GetGroupsFromPathArray(paths)

	groupPresent := false
	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		if formatted == groupName {
			groupPresent = true
			log.Debugf("Found matching Group")
		}
	}

	var channels []string

	if groupPresent {
		channels = tdms.GetChannelsFromPathArray(paths, groupName)
	} else {
		fmt.Println("File does not contain group named: ", groupName)
		log.Fatal("Invalid Group Name")
	}

	channelPresent := false
	for _, channel := range channels {
		formatted := strings.Replace(channel, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		if formatted == channelName {
			channelPresent = true
			log.Debugf("Found matching channel")
		}
	}

	var properties tdms.Properties
	// Get All Property Maps from segmetns containing a match, keep overriding to keep last
	if channelPresent {
		for _, val := range segments {
			for path, propMap := range val.PropMap {
				if path == ("/'" + groupName + "'/'" + channelName + "'") {
					for _, propValue := range propMap {
						properties = append(properties, propValue)
					}
				}
			}
		}
	}

	sort.Sort(properties)

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	for _, val := range properties {
		fmt.Fprintf(writer, "%s\t%s\n", val.Name, val.StringValue)
	}
	writer.Flush()
}

func DisplayChannelData(file *os.File, groupName string, channelName string) {
	segments, props := tdms.ReadAllSegments(file)

	paths := tdms.ReadAllUniqueTDMSObjects(segments)

	groups := tdms.GetGroupsFromPathArray(paths)

	groupPresent := false
	var groupString string
	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		if formatted == groupName {
			groupPresent = true
			groupString = group
			log.Debugf("Found matching Group")
		}
	}

	var channels []string
	var channelString string

	if groupPresent {
		channels = tdms.GetChannelsFromPathArray(paths, groupName)
	} else {
		fmt.Println("File does not contain group named: ", groupName)
		log.Fatalln("Invalid Group Name")
	}

	channelPresent := false
	for _, channel := range channels {
		formatted := strings.Replace(channel, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		if formatted == channelName {
			channelPresent = true
			channelString = channel
			log.Debugf("Found matching channel")
		}
	}

	if channelPresent {
		fullPath := groupString + "/" + channelString
		DisplayChannelRawData(file, fullPath, -1, 0, segments, props)
	}
}

