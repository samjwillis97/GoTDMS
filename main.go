package main

import (
	"bytes"
	"encoding/binary"
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
	objectOrder							 []string
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

const version string = "0.0.1.0"

func main() {
	var debug bool
	var help bool
	var json bool

	//TODO: Implement JSON Outputs

	flag.BoolVar(&json, "json", false, "Output in JSON Format")
	flag.BoolVar(&debug, "d", false, "Output Debug Log")
	flag.BoolVar(&debug, "debug", false, "Output Debug Log")
	flag.BoolVar(&help, "h", false, "Help")
	flag.BoolVar(&help, "help", false, "Help")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0)
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
			log.Debugln(len(os.Args))
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
				displayTDMSFile(file)
				file.Close()
			}
		}
	default:
		fmt.Println("Unkown Command")
		printHelp()
	}
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
	fmt.Println(version)
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("    help")
	fmt.Println("    list")
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  --debug, -d")
	fmt.Println("  --help, -h")
	fmt.Println("  --json")
	fmt.Println("  --version, -v")
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

func displayTDMSFile(file *os.File) {
	// Get All Segments, Find all Non Duplicates
	// Get Each Group, Each Channel and All Properties
	segments := readAllTDMSSegments(file)

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
	writer.Flush()
}

func displayTDMSGroups(file *os.File) {
	segments := readAllTDMSSegments(file)

	paths := readAllUniqueTDMSObjects(segments)

	groups := getGroupsFromPathArray(paths)

	for _, group := range groups {
		formatted := strings.Replace(group, "/", "", -1)
		formatted = strings.Replace(formatted, "'", "", -1)
		fmt.Println(formatted)
	}
}

func displayTDMSGroupChannels(file *os.File, groupName string) {
	segments := readAllTDMSSegments(file)

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
	segments := readAllTDMSSegments(file)

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
	// Get All Objets from each Segments
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
func readAllTDMSSegments(file *os.File) []Segment {
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

	// Iterate through Segments
	for {
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
		newSegment := readTDMSSegment(file, int64(segmentPos), 0, prevSegment)

		segments = append(segments, newSegment)
		prevSegment = newSegment
		segmentPos = newSegment.nextSegPos

		if segmentPos >= uint64(fi.Size()) {
			break
		}
	}

	log.Debugln("Finished Reading TDMS Segments")

	return segments
}

// Reads a TDMS Segment
// Includes:
// - Lead In
// - Meta Data
// Data is written in Segments, every time data is appended to a TDMS, a new segment is created
// A segment consists of Lead In, Meta Data, and Raw Data.
// There are exceptions to the rules
// hence Different Groups when written after each other will be in different seg
func readTDMSSegment(file *os.File, offset int64, whence int, prevSegment Segment) Segment {
	startPos, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSLeadIn: ", err)
	}
	log.Debugf("Reading TDMS Segement starting at: %d", startPos)

	// Read TDMS Lead In
	// leadIn := readTDMSLeadIn(file, offset, whence)
	leadIn := readTDMSLeadIn(file, 0, 1)

	// Read TDMS Meta Data
	objMap, objOrder, propMap := readTDMSMetaData(file, 0, 1, leadIn, prevSegment)

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

// Reads the TDMS Lead-In (28 Bytes)
// Includes:
// - Start Tag = 4 Bytes
// - ToC BitMask = 4 Bytes
// - Version Number = 4 Bytes
// - Segment Length = 8 Bytes
// - Metadata Length = 8 Bytes
// Total 28 Bytes
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func readTDMSLeadIn(file *os.File, offset int64, whence int) LeadInData {
	segmentStartPos, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSLeadIn: ", err)
	}

	log.Debugln("READING LEAD-IN")

	// Starts with a 4-byte tag that identifies a TDMS Segment ("TDSm")
	segStartTag := make([]byte, 4)
	_, err = io.ReadFull(file, segStartTag)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in readTDMSLeadIn: ", err)
	}
	if string(segStartTag) != "TDSm" {
		log.Fatal("Segment is not a TDMS")
	}
	log.Debugln("Valid TDMS Segment Starting at: ", segmentStartPos)

	// 4 Byte ToC BitMask
	// Example
	// Binary (Hexadecimal)		= 0E 00 00 00
	// ToC Mask								= 0x0000000E = 0b1110 = 0001 0001 0001 0000
	// Segment Contains: Object List, Meta Data, Raw Data
	tocBitMaskBytes := make([]byte, 4)
	_, err = io.ReadFull(file, tocBitMaskBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in readTDMSLeadIn: ", err)
	}
	tocBitMask := binary.LittleEndian.Uint32(tocBitMaskBytes)
	log.Debugln("ToC BitMask: ", tocBitMask)
	if (kTocMetaData & tocBitMask) == kTocMetaData {
		log.Debugln("Segment Contains Meta Data")
	}
	if (0b1000 & tocBitMask) == 0b1000 {
		log.Debugln("Segment Contains Raw Data")
	}
	if (0b10000000 & tocBitMask) == 0b10000000 {
		log.Debugln("Segment Contains DAQmx Raw Data")
	}
	if (0b100000 & tocBitMask) == 0b100000 {
		log.Debugln("Segment Contains Interleaved Data")
	}
	if (0b1000000 & tocBitMask) == 0b1000000 {
		log.Debugln("Segment Contains Big Endian Data")
	}
	if (0b100 & tocBitMask) == 0b100 {
		log.Debugln("Segment Contains New Object List")
	}

	// 4 Byte Version Number
	// 4713 = v2.0
	// 4712 = Older
	versionNumber := uint32FromTDMS(file, 0, 1)
	log.Debugln("Version Number: ", versionNumber)

	// 8 Bytes - Length of Remaining Segment
	// Also known as Next Segment Offset
	// Remaining Length = Overall Length of Segment - Length of Lead in ()
	// If an application encounters a problem writing, all bytes will = 0xFF
	// can only happen at EOF
	segLength := uint64FromTDMS(file, 0, 1)
	log.Debugln("Segment Length: ", segLength)

	// 8 Bytes - Length of Metadata in Segment
	// Also known as raw data offset
	// If segment contains no metadata will = 0
	metaLength := uint64FromTDMS(file, 0, 1)
	log.Debugln("Metadata Length: ", metaLength)

	leadInSize := uint64(28)

	nextSegPos := uint64(0)
	if segLength == 0xFFFFFFFFFFFFFFFF {
		log.Debugf("Segment incomplete, attempting to Read")
		fileStat, err := file.Stat()
		if err != nil {
			log.Fatal("Error return by file.Stat() in readTDMSLeadIn: ", err)
		}
		nextSegPos = uint64(fileStat.Size())
	} else {
		nextSegPos = uint64(segmentStartPos) + segLength + leadInSize
	}

	dataPos := uint64(segmentStartPos) + leadInSize + metaLength

	return LeadInData{
		tocBitMask,
		versionNumber,
		segLength,
		metaLength,
		nextSegPos,
		dataPos,
	}
}

// Read the TDMS MetaData
// Includes:
// - Reading Number of Objects in Segment
// - Object Paths
// - Object Info
// - Object Properties
//
// Starts at Byte Defined by Offset
// Whence is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func readTDMSMetaData(file *os.File, offset int64, whence int, leadin LeadInData, prevSegment Segment) (map[string]SegmentObject, []string, map[string]map[string]Property) {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Initialize Empty Map for Objects
	// TODO: Change to a Slice with a Map for Lookup
	objMap := make(map[string]SegmentObject)
	objOrder := make([]string, 0)

	// Init Map of Property Maps
	propertyMap := make(map[string]map[string]Property)

	// True if no MetaData
	if (kTocMetaData & leadin.ToCMask) != kTocMetaData {
		log.Debugln("Reuse Previous Segment Metadata")
		return prevSegment.objects, prevSegment.objectOrder, prevSegment.propMap
	}

	// TODO: Big Endianness with TocMask

	prevSegObjectNum := len(prevSegment.objects)

	if ((kTocNewObjList & leadin.ToCMask) == kTocNewObjList) || prevSegObjectNum == 0 {
	} else {
		// There can be a list of new objects that are appended,
		// or previous objects that are repeated with changed properties
		objMap = prevSegment.objects
		objOrder = prevSegment.objectOrder
	}

	log.Debugln("READING METADATA")

	// First 4 Bytes have number of objects in metadata
	numObjects := uint32FromTDMS(file, 0, 1)
	log.Debugln("Number of Objects: ", numObjects)

	// ar objects = make([]string, numObjects)
	for i := uint32(0); i < numObjects; i++ {
		log.Debugf("Reading Object %d \n", i)

		// Read Object Path
		objPath := stringFromTDMS(file, 0, 1)
		log.Debugf("Object %d Path: %s\n", i, objPath)

		// Read Raw Data Index/Length of Index Information
		// FF FF FF FF means there is no raw data
		// 69 12 00 00 DAQmx Format Changing Scaler
		// 69 13 00 00 DAQmx Digital Line Scaler
		// Matches Previous Segment Same Object i.e. use previous
		// Otherwise
		rawDataIndexHeaderBytes := make([]byte, 4)
		_, err := io.ReadFull(file, rawDataIndexHeaderBytes)
		if err != nil {
			log.Fatal("Error return from io.ReadFull in readTDMSObject: ", err)
		}
		log.Debugf("Object Raw Data Index: % x", rawDataIndexHeaderBytes)

		// check if we already have objPath
		// only executes if present true, present will be true if found in map
		// first value is the value found
		if val, present := objMap[objPath]; present {
			log.Debugf("Updating Existing Object\n")
			// Current Header says No Data
			if bytes.Compare(rawDataIndexHeaderBytes, noRawDataValue) == 0 {
				// Matched Header says Has Data
				if bytes.Compare(val.rawDataIndexHeader, noRawDataValue) != 0 {
					objMap[objPath] = SegmentObject{
						rawDataIndexHeaderBytes,
						val.rawDataIndex,
					}
				}
				// Current Header Matches Previous
			} else if bytes.Compare(rawDataIndexHeaderBytes, matchesPreviousValue) == 0 {
				// Previous has No Raw Data
				if bytes.Compare(val.rawDataIndexHeader, noRawDataValue) == 0 {
					objMap[objPath] = SegmentObject{
						rawDataIndexHeaderBytes,
						val.rawDataIndex,
					}
				}
				// New Segment Metadata OR Updates to Existing Data
			} else {
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					readTDMSRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
				}
			}
		} else if val, present := prevSegment.objects[objPath]; present {
			log.Debugf("Reusing Previous Segment Object\n")
			// reuse previous object
			if bytes.Compare(rawDataIndexHeaderBytes, noRawDataValue) == 0 {
				// Reuse Segment  But Leave Data Index Information as Set Previously
				if bytes.Compare(val.rawDataIndexHeader, noRawDataValue) == 0 {
					// Previous Segment has Data
					// Copy Previos to Current, Leaving Header
					objMap[objPath] = SegmentObject{
						rawDataIndexHeaderBytes,
						val.rawDataIndex,
					}
					objOrder = append(objOrder, objPath)
				} else {
					// Copy Completely
					objMap[objPath] = val
					objOrder = append(objOrder, objPath)
				}
				// Matches Previous
			} else if bytes.Compare(rawDataIndexHeaderBytes, matchesPreviousValue) == 0 {
				if bytes.Compare(val.rawDataIndexHeader, noRawDataValue) != 0 {
					// Copy Previos to Current, Leaving Header
					objMap[objPath] = SegmentObject{
						rawDataIndexHeaderBytes,
						val.rawDataIndex,
					}
				} else {
					// Copy Completely
					objMap[objPath] = val
				}
			} else {
				// Changed Metadata in this Section
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					readTDMSRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
				}
			}
		} else {
			log.Debugf("New Segment Object\n")
			// New Segment Object
			if bytes.Compare(rawDataIndexHeaderBytes, matchesPreviousValue) == 0 {
				log.Fatal("Raw Data Index says to reuse previous, though this object has not been seen before")
			} else if bytes.Compare(rawDataIndexHeaderBytes, noRawDataValue) != 0 {
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					readTDMSRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
				}
				objOrder = append(objOrder, objPath)
			} else {
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					RawDataIndex{
						0,
						0,
						0,
						0,
					},
				}
				objOrder = append(objOrder, objPath)
			}
		}

		// Number of Object Properties
		numProperties := uint32FromTDMS(file, 0, 1)
		log.Debugf("Number of Object %d Properties: %d\n", i, numProperties)

		// Read Properties
		for j := uint32(0); j < numProperties; j++ {
			log.Debugf("Reading Object %d Property %d\n", i, j)
			property := readTDMSProperty(file, 0, 1)
			// if propMap, present := propertyMap[objPath]; present {
			if _, present := propertyMap[objPath]; present {
				// Property Maps Exists for Path
				propertyMap[objPath][property.name] = property
				// if _, present := propMap[property.name]; present {
				// 	// Property Exists in Property Map
				// 	// Update it
				// 	// MAYBE NOT NECESSARY?
				// 	propertyMap[objPath][property.name] = property
				// } else {
				// 	propertyMap[objPath][property.name] = property
				// }
			} else {
				// Property Map Doesn't exist for Path yet
				initMap := map[string]Property{
					property.name: {
						property.name,
						property.dataType,
						property.valuePosition,
						property.stringValue,
					},
				}
				propertyMap[objPath] = initMap
			}
		}
	}

	//TODO: REMOVE
	// fmt.Println()
	// fmt.Println("MAP")
	// fmt.Println(objMap)
	// fmt.Println()
	// fmt.Println("PROPERTIES")
	// fmt.Println(propertyMap)

	return objMap, objOrder, propertyMap
}

// Read the Properties for a TDMS Object
// TODO
// Change this to output a list of properties
func readTDMSProperty(file *os.File, offset int64, whence int) Property {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Property Name
	propertyName := stringFromTDMS(file, 0, 1)
	// log.Debugf("Property Name: %s\n", propertyName)

	// Debuged in Hex
	propertyDataType := uint32FromTDMS(file, 0, 1)
	propertyTdsDataType := tdsDataType(propertyDataType)

	// Position for reading later
	valuePosition, _ := file.Seek(0, 1)

	// Property Value coerced to String
	var valueString string

	// TODO
	// Finish This
	switch propertyTdsDataType {
	default:
		log.Fatal("Property Data Type Unkown")
	case String:
		valueString = stringFromTDMS(file, 0, 1)
	case Int32:
		valueString = fmt.Sprintf("%d", int32FromTDMS(file, 0, 1))
	case Uint32:
		valueString = fmt.Sprintf("%d", uint32FromTDMS(file, 0, 1))
	case Uint64:
		valueString = fmt.Sprintf("%d", uint64FromTDMS(file, 0, 1))
	case DBL:
		valueString = fmt.Sprintf("%e", DBLFromTDMS(file, 0, 1))
	case Timestamp:
		valueString = timeFromTDMS(file, 0, 1).String()
	}

	return Property{
		propertyName,
		propertyTdsDataType,
		valuePosition,
		valueString,
	}
}

func readTDMSRawDataIndex(file *os.File, offset int64, whence int, rawDataIndexHeader []byte) RawDataIndex {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return by file.Seek in readTDMSRawDataIndex: ", err)
	}

	indexLength := binary.LittleEndian.Uint32(rawDataIndexHeader)
	log.Debugf("Object Index Length: %d\n", indexLength)

	dataType := tdsDataType(uint32FromTDMS(file, 0, 1))
	log.Debugf("Object Data Type: %d\n", dataType)

	// must equal 1 for v2.0
	arrayDimension := uint32FromTDMS(file, 0, 1)
	if arrayDimension != 1 {
		log.Fatal("Not Valid TDMS 2.0, Data Dimension is not 1")
	}

	numValues := uint64FromTDMS(file, 0, 1)
	log.Debugf("Object Number of Values: %d\n", numValues)

	dataSize := 0
	switch tdsDataType(dataType) {
	case Int8, Uint8, Boolean:
		dataSize = 1
	case Int16, Uint16:
		dataSize = 2
	case Int32, Uint32, SGL, SGLwUnit:
		dataSize = 4
	case Int64, Uint64, DBL, DBLwUnit:
		dataSize = 8
	case Timestamp:
		dataSize = 16
	}

	channelRawDataSize := uint64(dataSize) * uint64(arrayDimension) * numValues
	log.Debugf("Channel Raw Data Size: %d\n", channelRawDataSize)

	return RawDataIndex{
		tdsDataType(dataType),
		arrayDimension,
		numValues,
		channelRawDataSize,
	}
}

// REQUIRES
// ObjMap/Segment.objects
// segment.nextSegPos
// segment.dataPos
func calculateChunks(objects map[string]SegmentObject, nextSegPos uint64, dataPos uint64) uint64 {
	dataSize := uint64(0)
	for _, e := range objects {
		dataSize += e.rawDataIndex.rawDataSize
	}
	log.Debugf("Data Size: %d", dataSize)

	totalDataSize := nextSegPos - dataPos
	log.Debugf("Total Data Size: %d", totalDataSize)

	if dataSize < 0 || totalDataSize < 0 {
		log.Fatal("Negative data size")
	} else if dataSize == 0 {
		// npTDMS: sometimes kTocRawData is set, but there isn't actually any data
		if totalDataSize != dataSize {
			log.Fatal("Zero channel data size but data length")
		}
		numChunks := uint64(0)
		return numChunks
	}

	// Checking for Multiple
	chunkRemainder := totalDataSize % dataSize
	if chunkRemainder == 0 {
		numChunks := uint64(totalDataSize / dataSize)
		return numChunks
	} else {
		log.Fatal("Data Size is not a multiple of Chunk Size")
		return uint64(0)
	}
}

// TODO
// Change these to read one or more
// Probably more efficient

// Reads an int32 from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func int32FromTDMS(file *os.File, offset int64, whence int) int32 {
	value := uint32FromTDMS(file, offset, whence)
	return int32(value)
}

// Reads an int64 from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func int64FromTDMS(file *os.File, offset int64, whence int) int64 {
	value := uint64FromTDMS(file, offset, whence)
	return int64(value)
}

// Reads a string from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func stringFromTDMS(file *os.File, offset int64, whence int) string {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in stringFromTDMS: ", err)
	}
	// Get Length of String
	// Required to be in the first 4 bytes
	stringLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(file, stringLengthBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in stringFromTDMS: ", err)
	}
	stringLength := binary.LittleEndian.Uint32(stringLengthBytes)

	// Get String Bytes
	stringBytes := make([]byte, stringLength)
	_, err = io.ReadFull(file, stringBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in stringFromTDMS: ", err)
	}

	return string(stringBytes)
}

// Reads a uint32 from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func uint32FromTDMS(file *os.File, offset int64, whence int) uint32 {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in uint32FromTDMS: ", err)
	}

	intBytes := make([]byte, 4)
	_, err = io.ReadFull(file, intBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in uint32FromTDMS: ", err)
	}
	intNumber := binary.LittleEndian.Uint32(intBytes)

	return intNumber
}

func uint32ArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []uint32 {
	size := int64(4)

	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in uint32FromTDMS: ", err)
	}

	intByteArray := make([]byte, number*size)
	_, err = io.ReadFull(file, intByteArray)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in uint32FromTDMS: ", err)
	}

	var vals []uint32

	for i := int64(0); i < number; i++ {
		startBit := i * size
		endBit := startBit + size
		val := binary.LittleEndian.Uint32(intByteArray[startBit:endBit])
		vals = append(vals, val)
	}

	return vals
}

func uint64ArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []uint64 {
	size := int64(8)

	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in uint64FromTDMS: ", err)
	}

	intByteArray := make([]byte, number*size)
	_, err = io.ReadFull(file, intByteArray)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in uint64FromTDMS: ", err)
	}

	var vals []uint64

	for i := int64(0); i < number; i++ {
		startBit := i * size
		endBit := startBit + size
		val := binary.LittleEndian.Uint64(intByteArray[startBit:endBit])
		vals = append(vals, val)
		// log.Debugf("Value %d: %d\n", i, val)
	}

	return vals
}

func DBLArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []float64 {
	size := int64(8)

	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in DBLArrayFromTDMS: ", err)
	}

	intByteArray := make([]byte, number*size)
	_, err = io.ReadFull(file, intByteArray)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in DBLArrayFromTDMS: ", err)
	}

	var vals []float64

	for i := int64(0); i < number; i++ {
		startBit := i * size
		endBit := startBit + size
		val := math.Float64frombits(binary.LittleEndian.Uint64(intByteArray[startBit:endBit]))
		vals = append(vals, val)
		// log.Debugf("Value %d: %.2f\n", i, val)
	}

	return vals
}

// Reads a uint64 from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func uint64FromTDMS(file *os.File, offset int64, whence int) uint64 {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in uint64FromTDMS: ", err)
	}

	intBytes := make([]byte, 8)
	_, err = io.ReadFull(file, intBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in uint64FromTDMS: ", err)
	}
	intNumber := binary.LittleEndian.Uint64(intBytes)

	return intNumber
}

// Reads a DBL from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func DBLFromTDMS(file *os.File, offset int64, whence int) float64 {
	value := uint64FromTDMS(file, offset, whence)
	return math.Float64frombits(value)
}

// Reads a Timestamp from a TDMS File
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
func timeFromTDMS(file *os.File, offset int64, whence int) time.Time {
	posFractions := uint64FromTDMS(file, offset, whence)
	LVseconds := int64FromTDMS(file, 0, 1)
	nanoSeconds := float64(posFractions) * math.Pow(2, -64) * 1e9
	secondsToUnix := 2.083e9
	timeValue := time.Unix(0, 0)
	if LVseconds != 0 && nanoSeconds != 0 {
		timeValue = time.Unix(LVseconds-int64(secondsToUnix), int64(nanoSeconds))
	}
	return timeValue
}

func minMaxFloat64Slice(y []float64) (min float64, max float64) {
	min = y[0]
	max = y[0]

	for _, v := range y {
		if v > max {
			max = v
		} else if v < min {
			min = v
		}
	}

	return min, max
}

func averageFloat64Slice(y []float64) (avg float64) {
	avg = y[0]

	for _, v := range y {
		avg += v
	}

	return avg / float64(len(y))
}

// Sorting for Properties
func (p Properties) Len() int { return len(p) }

func (p Properties) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p Properties) Less(i, j int) bool {
	var si string = p[i].name
	var sj string = p[j].name
	var si_low = strings.ToLower(si)
	var sj_low = strings.ToLower(sj)
	if si_low == sj_low {
		return si < sj
	}
	return si_low < sj_low
}
