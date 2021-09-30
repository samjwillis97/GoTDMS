package cli

import (
	"fmt"
	"os"
	"strings"
	"sort"
	"text/tabwriter"

	"github.com/samjwillis97/GoTDMS/pkg/tdms"
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
				for path, propMap := range val.propMap {
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

					fmt.Fprintf(writer, propFormatString.String(), val.name, val.stringValue)
				}
			}
		}
	}
	writer.Flush()
}
