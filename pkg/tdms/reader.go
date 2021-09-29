package tdms

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Each Segment of a TDMS Has a Lead In Section
type LeadInData struct {
	ToCMask       uint32
	versionNumber uint32
	nextSegOffset uint64
	rawDataOffset uint64
	nextSegPos    uint64
	dataPos       uint64
}

// Each Segment of a TDMS Consists of all this information
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

// Required Data for each Object in a Segment
type SegmentObject struct {
	rawDataIndexHeader []byte
	rawDataIndex       RawDataIndex
}

// Information from Raw Data Index
type RawDataIndex struct {
	dataType       tdsDataType
	arrayDimension uint32
	numValues      uint64
	rawDataSize    uint64
}

type tdsDataType uint64

type Property struct {
	name          string
	dataType      tdsDataType
	valuePosition int64
	stringValue   string
}

type Properties []Property

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

// Constants

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

// Get All Segments of TDMS File
func readAllSegments(file *os.File) ([]Segment, map[string]map[string]Property) {
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
		newSegment := readSegment(file, int64(segmentPos), 0, prevSegment, allPrevSegObjs)

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
func readSegment(file *os.File, offset int64, whence int, prevSegment Segment, allPrevSegObjs map[string]SegmentObject) Segment {
	startPos, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSLeadIn: ", err)
	}
	log.Debugf("Reading TDMS Segement starting at: %d", startPos)

	// Read TDMS Lead In
	// leadIn := readTDMSLeadIn(file, offset, whence)
	leadIn := readLeadIn(file, 0, 1)

	// Read TDMS Meta Data
	objMap, objOrder, propMap := readMetaData(file, 0, 1, leadIn, prevSegment, allPrevSegObjs)

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

// Reads the TDMS Lead-In (28 Bytes) of a Segment
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
//
// Returns LeadInData
func readLeadIn(file *os.File, offset int64, whence int) LeadInData {
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
	versionNumber := readUint32(file, 0, 1)
	log.Debugln("Version Number: ", versionNumber)

	// 8 Bytes - Length of Remaining Segment
	// Also known as Next Segment Offset
	// Remaining Length = Overall Length of Segment - Length of Lead in ()
	// If an application encounters a problem writing, all bytes will = 0xFF
	// can only happen at EOF
	segLength := readUint64(file, 0, 1)
	log.Debugln("Segment Length: ", segLength)

	// 8 Bytes - Length of Metadata in Segment
	// Also known as raw data offset
	// If segment contains no metadata will = 0
	metaLength := readUint64(file, 0, 1)
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

// Read the TDMS MetaData of a Segment
//
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
//
// Returns Segment Objects and Properties
func readMetaData(file *os.File, offset int64, whence int, leadin LeadInData, prevSegment Segment, allPrevSegObjs map[string]SegmentObject) (map[string]SegmentObject, []string, map[string]map[string]Property) {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Initialize Empty Map for Objects
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
	numObjects := readUint32(file, 0, 1)
	log.Debugln("Number of Objects: ", numObjects)

	// ar objects = make([]string, numObjects)
	for i := uint32(0); i < numObjects; i++ {
		log.Debugf("Reading Object %d \n", i)

		// Read Object Path
		objPath := readString(file, 0, 1)
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

		// check if we already have objPath in Map
		// only executes if present true, present will be true if found in map
		// first value is the value found
		if val, present := objMap[objPath]; present {
			log.Debugf("Updating Existing %s Object\n", objPath)
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
					readRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
				}
			}
		} else if val, present := allPrevSegObjs[objPath]; present {
			log.Debugf("Reusing Previous %s Object\n", objPath)
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
					objOrder = append(objOrder, objPath)
				} else {
					// Copy Completely
					objMap[objPath] = val
					objOrder = append(objOrder, objPath)
				}
			} else {
				// Changed Metadata in this Section
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					readRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
				}
				objOrder = append(objOrder, objPath)
			}
		} else {
			log.Debugf("New Segment Object: %s\n", objPath)
			// New Segment Object
			if bytes.Compare(rawDataIndexHeaderBytes, matchesPreviousValue) == 0 {
				log.Fatalln("Raw Data Index says to reuse previous, though this object has not been seen before: ", objPath)
			} else if bytes.Compare(rawDataIndexHeaderBytes, noRawDataValue) != 0 {
				objMap[objPath] = SegmentObject{
					rawDataIndexHeaderBytes,
					readRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes),
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
		numProperties := readUint32(file, 0, 1)
		log.Debugf("Number of Object %d Properties: %d\n", i, numProperties)

		// Read Properties
		for j := uint32(0); j < numProperties; j++ {
			log.Debugf("Reading Object %d Property %d\n", i, j)
			property := readProperty(file, 0, 1)
			// if propMap, present := propertyMap[objPath]; present {
			if _, present := propertyMap[objPath]; present {
				// Property Maps Exists for Path
				propertyMap[objPath][property.name] = property
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
	return objMap, objOrder, propertyMap
}

// Reads Raw Data Index of a Segment Object
//
// Returns RawDataIndex
func readRawDataIndex(file *os.File, offset int64, whence int, rawDataIndexHeader []byte) RawDataIndex {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return by file.Seek in readTDMSRawDataIndex: ", err)
	}

	indexLength := binary.LittleEndian.Uint32(rawDataIndexHeader)
	log.Debugf("Object Index Length: %d\n", indexLength)

	dataType := tdsDataType(readUint32(file, 0, 1))
	log.Debugf("Object Data Type: %d\n", dataType)

	// must equal 1 for v2.0
	arrayDimension := readUint32(file, 0, 1)
	if arrayDimension != 1 {
		log.Fatal("Not Valid TDMS 2.0, Data Dimension is not 1")
	}

	numValues := readUint64(file, 0, 1)
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

// Reads a single property from a Segment Object
func readProperty(file *os.File, offset int64, whence int) Property {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Property Name
	propertyName := readString(file, 0, 1)
	// log.Debugf("Property Name: %s\n", propertyName)

	// Debuged in Hex
	propertyDataType := readUint32(file, 0, 1)
	propertyTdsDataType := tdsDataType(propertyDataType)

	// Position for reading later
	valuePosition, _ := file.Seek(0, 1)

	// Property Value coerced to String
	var valueString string

	// TODO
	// Finish This
	// Converts to Properties
	switch propertyTdsDataType {
	default:
		log.Fatal("Property Data Type Unkown")
	case String:
		valueString = readString(file, 0, 1)
	case Int32:
		valueString = fmt.Sprintf("%d", readInt32(file, 0, 1))
	case Uint32:
		valueString = fmt.Sprintf("%d", readUint32(file, 0, 1))
	case Uint64:
		valueString = fmt.Sprintf("%d", readUint64(file, 0, 1))
	case DBL:
		valueString = fmt.Sprintf("%e", readDBL(file, 0, 1))
	case Timestamp:
		valueString = readTime(file, 0, 1).String()
	}

	return Property{
		propertyName,
		propertyTdsDataType,
		valuePosition,
		valueString,
	}
}
