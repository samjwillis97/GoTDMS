package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	"github.com/samjwillis97/GoTDMS/pkg/analysis"
)

type LeadInData struct {
	ToCMask       uint32
	versionNumber uint32
	nextSegOffset uint64
	rawDataOffset uint64
	nextSegPos    uint64
	dataPos       uint64
}

type SegmentObject struct {
	rawDataIndexHeader []byte
	rawDataIndex       RawDataIndex
}

type RawDataIndex struct {
	dataType       tdsDataType
	arrayDimension uint32
	numValues      uint64
	rawDataSize    uint64
}

type Segment struct {
	position                 uint64
	numChunks                uint64
	objects                  map[string]SegmentObject
	objectOrder              []string
	kToCMask                 uint32
	nextSegPos               uint64
	dataPos                  uint64
	finalChunkLengthOverride uint64
	objectIndex              uint64
	propMap                  map[string]map[string]Property
}

type Property struct {
	name          string
	dataType      tdsDataType
	valuePosition int64
	stringValue   string
}

type Properties []Property

type tdsDataType uint64

const (
	Void       tdsDataType = 0
	Int8       tdsDataType = 1
	Int16      tdsDataType = 2
	Int32      tdsDataType = 3
	Int64      tdsDataType = 4
	Uint8      tdsDataType = 5
	Uint16     tdsDataType = 6
	Uint32     tdsDataType = 7
	Uint64     tdsDataType = 8
	SGL        tdsDataType = 9
	DBL        tdsDataType = 10
	EXT        tdsDataType = 11
	SGLwUnit   tdsDataType = 0x19
	DBLwUnit   tdsDataType = 0x1A
	EXTwUnit   tdsDataType = 0x1B
	String     tdsDataType = 0x20
	Boolean    tdsDataType = 0x21
	Timestamp  tdsDataType = 0x44
	ComplexSGL tdsDataType = 0x08000C
	ComplexDBL tdsDataType = 0x10000D
	DAQmx      tdsDataType = 0xFFFFFF
)

const (
	kTocMetaData        uint32 = 0x2
	kTocRawData         uint32 = 0x8
	kTocDAQmxRawData    uint32 = 0x80
	kTocInterleavedData uint32 = 0x20
	kTocBigEndian       uint32 = 0x40
	kTocNewObjList      uint32 = 0x4
)

var (
	noRawDataValue            = []byte{255, 255, 255, 255}
	matchesPreviousValue      = []byte{0, 0, 0, 0}
	daqmxFormatChangingScaler = []byte{69, 12, 00, 00}
	daqmxDigitalLineScaler    = []byte{69, 13, 00, 00}
)

const versionString string = "0.0.1.0"

func main() {
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
					displayTDMSGroups(file)
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
					displayTDMSGroupChannels(file, groupName)
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
					displayTDMSChannelProperties(file, groupName, channelName)
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
				displayTDMSFile(file, verbose)
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
					displayChannelData(file, groupName, channelName)
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

func displayTDMSFile(file *os.File, verbose bool) {
	// Get All Segments, Find all Non Duplicates
	// Get Each Group, Each Channel and All Properties
	segments, _ := readAllTDMSSegments(file)

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

func displayChannelData(file *os.File, groupName string, channelName string) {
	segments, props := readAllTDMSSegments(file)

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
		displayChannelRawData(file, fullPath, -1, 0, segments, props)
	}
}

func displayTDMSGroups(file *os.File) {
	segments, _ := readAllTDMSSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		fmt.Println(formatted)
	}
}

func displayTDMSGroupChannels(file *os.File, groupName string) {
	segments, _ := readAllTDMSSegments(file)

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

func displayTDMSChannelProperties(file *os.File, groupName string, channelName string) {
	segments, _ := readAllTDMSSegments(file)

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

func getGroupsFromPathArray(paths []string) []string {
	var groups []string
	for _, path := range paths {
		if (path != "/") && (len(strings.Split(path, "/")) == 2) {
			groups = append(groups, path)
		}
	}
	return groups
}

func readAllUniqueTDMSObjects(segments []Segment) []string {
	// Get All Objets from each Segment
	// Remove all duplicates
	// Remove "/"
	// Remove "/<string>/<string>"
	// Effectively using a map as a set
	pathSet := make(map[string]bool)
	pathArray := make([]string, 0)
	for _, seg := range segments {
		for _, path := range seg.objectOrder {
			exists := pathSet[path]
			if !(exists) {
				pathSet[path] = true
				pathArray = append(pathArray, path)
			}
		}
	}

	return pathArray
}

func getChannelsFromPathArray(paths []string, group string) []string {
	var channels []string
	for _, path := range paths {
		splitString := strings.Split(path, "/")
		if (path != "/") && (len(splitString) == 3) && splitString[1] == ("'"+group+"'") {
			channels = append(channels, splitString[2])
		}
	}
	return channels
}

// Get All Segments of TDMS File
func readAllTDMSSegments(file *os.File) ([]Segment, map[string]map[string]Property) {
	// Get File Size
	fi, err := file.Stat()
	if err != nil {
		log.Fatal("Could not Obtain File Stats: ", err)
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal("Error return from file.Seek in readAllTDMSSegments: ", err)
	}

	// Init Variables
	var segments []Segment
	segmentPos := uint64(0)
	allPrevSegObjs := make(map[string]SegmentObject)

	prevSegment := Segment{
		0,
		0,
		map[string]SegmentObject{},
		[]string{},
		0,
		0,
		0,
		0,
		0,
		map[string]map[string]Property{},
	}

	// Iterate through Segments
	for {
		newSegment := readTDMSSegment(file, int64(segmentPos), 0, prevSegment, allPrevSegObjs)

		segments = append(segments, newSegment)
		prevSegment = newSegment
		segmentPos = newSegment.nextSegPos

		for path, val := range newSegment.objects {
			allPrevSegObjs[path] = val
		}

		if segmentPos >= uint64(fi.Size()) {
			break
		}
	}

	// TODO:
	// Iterate through all Each Segments Properties, only keeping latest
	// Return the latest Properties
	objProperties := make(map[string]map[string]Property, 0)
	for _, seg := range segments {
		for path, propMap := range seg.propMap {
			_, pathPresent := objProperties[path]
			if !pathPresent {
				objProperties[path] = propMap
			} else {
				for prop, propVals := range propMap {
					objProperties[path][prop] = propVals
				}
			}
		}
	}

	log.Debugln("Finished Reading TDMS Segments")

	return segments, objProperties
}

// Reads a TDMS Segment
// Includes:
// - Lead In
// - Meta Data
// Data is written in Segments, every time data is appended to a TDMS, a new segment is created
// A segment consists of Lead In, Meta Data, and Raw Data.
// There are exceptions to the rules
// hence Different Groups when written after each other will be in different seg
func readTDMSSegment(file *os.File, offset int64, whence int, prevSegment Segment, allPrevSegObjs map[string]SegmentObject) Segment {
	startPos, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSLeadIn: ", err)
	}
	log.Debugf("Reading TDMS Segement starting at: %d", startPos)

	// Read TDMS Lead In
	// leadIn := readTDMSLeadIn(file, offset, whence)
	leadIn := readTDMSLeadIn(file, 0, 1)

	// Read TDMS Meta Data
	objMap, objOrder, propMap := readTDMSMetaData(file, 0, 1, leadIn, prevSegment, allPrevSegObjs)

	// Calculate Number of Chunks
	numChunks := calculateChunks(objMap, leadIn.nextSegPos, leadIn.dataPos)

	// Object Index
	index := prevSegment.objectIndex + 1

	// TODO: Finish Reading Raw Data
	// if (0b100000 & leadIn.ToCMask) == 0b100000 {
	// Segment Contains Interleaved Data
	// }

	// Read Data Ch by Ch
	// for key, element := range objMap {
	// 	switch element.dataType {
	// 	default:
	// 		_, err := file.Seek(int64(element.rawDataSize), 1)
	// 		if err != nil {
	// 			log.Fatal("Error return by file.Seek in readTDMSSegment: ", err)
	// 		}
	// 	case DBL:
	// 		data := DBLArrayFromTDMS(file, int64(element.numValues), 0, 1)
	// 		dataMin, dataMax := minMaxFloat64Slice(data)
	// 		log.Debugf("Read %s Values\n", key)
	// 		log.Debugf("Number of Values: %d\n", len(data))
	// 		log.Debugf("Max Value: %.6f\n", dataMax)
	// 		log.Debugf("Min Value: %.6f\n", dataMin)
	// 		log.Debugf("Average Value: %.6f\n", averageFloat64Slice(data))
	// 	}
	// }
	return Segment{
		uint64(startPos),
		numChunks,
		objMap,
		objOrder,
		leadIn.ToCMask,
		leadIn.nextSegPos,
		leadIn.dataPos,
		0, //TODO: Implement
		index,
		propMap,
	}
}



func displayChannelRawData(file *os.File, channelPath string, length int64, offset uint64, allSegments []Segment, allProps map[string]map[string]Property) {
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

					wf_increment := tdms.DBLFromTDMS(file, allProps[objPath]["wf_increment"].valuePosition, 0)
					wf_samples := tdms.Int32FromTDMS(file, allProps[objPath]["wf_samples"].valuePosition, 0)
					wf_start_time := tdms.TimeFromTDMS(file, allProps[objPath]["wf_start_time"].valuePosition, 0)

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
						dataSGL := tdms.SGLArrayFromTDMS(file, int64(obj.rawDataIndex.numValues), int64(obj.rawDataIndex.rawDataSize), 1)

						//convert data to float64
						for _, val := range dataSGL {
							data = append(data, float64(val))
						}

					case DBL:
						data = tdms.DBLArrayFromTDMS(file, int64(obj.rawDataIndex.numValues), int64(obj.rawDataIndex.rawDataSize), 1)

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
			} else {
			}
		}
	}
	writer.Flush()
}

