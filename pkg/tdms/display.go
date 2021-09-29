package tdms

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/samjwillis97/GoTDMS/pkg/analysis"
	log "github.com/sirupsen/logrus"
)

func DisplayFile(file *os.File, verbose bool) {
	// Get All Segments, Find all Non Duplicates
	// Get Each Group, Each Channel and All Properties
	segments, _ := readAllSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

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

		channels := getChannelsFromPathArray(paths, formattedG)

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

			var properties Properties
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

func DisplayChannelData(file *os.File, groupName string, channelName string) {
	segments, props := readAllSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

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
		channels = getChannelsFromPathArray(paths, groupName)
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
			channelString = channel
			log.Debugf("Found matching channel")
		}
	}

	if channelPresent {
		fullPath := groupString + "/" + channelString
		DisplayChannelRawData(file, fullPath, -1, 0, segments, props)
	}
}

func DisplayGroups(file *os.File) {
	segments, _ := readAllSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		fmt.Println(formatted)
	}
}

func DisplayGroupChannels(file *os.File, groupName string) {
	segments, _ := readAllSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

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
		channels = getChannelsFromPathArray(paths, groupName)
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
	segments, _ := readAllSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

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
		channels = getChannelsFromPathArray(paths, groupName)
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

	var properties Properties
	// Get All Property Maps from segmetns containing a match, keep overriding to keep last
	if channelPresent {
		for _, val := range segments {
			for path, propMap := range val.propMap {
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
		fmt.Fprintf(writer, "%s\t%s\n", val.name, val.stringValue)
	}
	writer.Flush()
}

func DisplayChannelRawData(file *os.File, channelPath string, length int64, offset uint64, allSegments []Segment, allProps map[string]map[string]Property) {
	// Determine Data Type of Segment
	// if TWF, defined by the properties
	// return RMS, P-P, CF for the whole file, add option for Block-by-block, that returns a slice
	firstSeg := true

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	// Iterate through all File Segments
	for i, segment := range allSegments {
		// Iterate through all the objects in order
		// Skipping over the data we don't need to read
		for _, objPath := range segment.objectOrder {
			obj := segment.objects[objPath]

			if objPath == channelPath {

				_, wfStartPresent := allProps[objPath]["wf_start_time"]
				_, wfStartOffsetPresent := allProps[objPath]["wf_start_offset"]
				_, wfIncrementPresent := allProps[objPath]["wf_increment"]
				_, wfSamplesPresent := allProps[objPath]["wf_samples"]

				// fmt.Println("New Segment:", i)
				// fmt.Println(allProps[objPath])

				if wfStartPresent && wfStartOffsetPresent && wfIncrementPresent && wfSamplesPresent {
					log.Debugln("Waveform Present")

					wf_increment := readDBL(file, allProps[objPath]["wf_increment"].valuePosition, 0)
					wf_samples := readInt32(file, allProps[objPath]["wf_samples"].valuePosition, 0)
					wf_start_time := readTime(file, allProps[objPath]["wf_start_time"].valuePosition, 0)

					if firstSeg {
						fmt.Printf("TDMS Path:\t%s\n", channelPath)
						fmt.Printf("Sample Rate:\t%d Hz\n", int(1/wf_increment))
						fmt.Printf("Channel Length:\t%d Samples\n", wf_samples)
						fmt.Printf("Start Time: \t%s\n", wf_start_time)
						fmt.Printf("Total Segments:\t%d\n", len(allSegments))

						fmt.Fprintf(writer, "\nSeg No. \tRMS \tP-P \tCF\n")
					}

					_, err := file.Seek(int64(segment.dataPos), 0)
					if err != nil {
						log.Fatalln("Error from file.Seek in readChannelRawData")
					}

					data := make([]float64, 0)

					switch obj.rawDataIndex.dataType {
					case SGL:
						dataSGL := readSGLArray(file, int64(obj.rawDataIndex.numValues), int64(obj.rawDataIndex.rawDataSize), 1)

						//convert data to float64
						for _, val := range dataSGL {
							data = append(data, float64(val))
						}

					case DBL:
						data = readDBLArray(file, int64(obj.rawDataIndex.numValues), int64(obj.rawDataIndex.rawDataSize), 1)

					default:
						log.Fatal("Data Type Not Implemented")
					}

					rms := analysis.RmsFloat64Slice(data)
					min, max := analysis.MinMaxFloat64Slice(data)
					pp := math.Abs(max - min)
					cf := max / rms

					fft := analysis.VibFFT(data, wf_increment, 0)

					fmt.Println(analysis.MaxFloat64(fft))
					fmt.Println()

					fmt.Fprintf(writer, "%d \t%.4f \t%.4f \t%.4f\n", i, rms, pp, cf)

					firstSeg = false
				}
			}
		}
	}
	writer.Flush()
}
