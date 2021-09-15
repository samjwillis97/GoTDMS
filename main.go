package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
	"time"
)

type LeadInData struct {
	ToCMask       uint32
	versionNumber uint32
	nextSegOffset uint64
	rawDataOffset uint64
	nextSegPos    uint64
	dataPos       uint64
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
	objects                  map[string]RawDataIndex
	kToCMask                 uint32
	nextSegPos               uint64
	dataPos                  uint64
	finalChunkLengthOverride uint64
	objectIndex              uint64
}

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

// kTocMetaData (1L << 1)					= 0b0000 0010			= 0x2
// kTocRawData (1L << 3)					= 0b0000 1000			= 0x8
// kTocDAQmxRawData (1L << 7)			= 0b1000 0000			= 0x80
// kToxInterleavedData (1L << 5)  = 0b0010 0000			= 0x20
// kTocBigEndian (1L << 6)				= 0b0100 0000			= 0x40
// kTocNewObjList (1L << 2)				= 0b0000 0100			= 0x4
const (
	kTocMetaData        uint32 = 0x2
	kTocRawData         uint32 = 0x8
	kTocDAQmxRawData    uint32 = 0x80
	kTocInterleavedData uint32 = 0x20
	kTocBigEndian       uint32 = 0x40
	kTocNewObjList      uint32 = 0x4
)

var (
	noRawDataValue = []byte{255, 255, 255, 255}
)

func main() {
	initLogging()

	file, err := os.OpenFile("testFiles/demo.tdms", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal("Error return from os.OpenFile: ", err)
	}

	// Reading a TDMS File
	// https://www.ni.com/en-au/support/documentation/supplemental/07/tdms-file-format-internal-structure.html
	// TODO: Proper Reading and More Error Checks with File Lengths
	// TODO: Workout and Implement Interfaces

	emptySegment := Segment{
		0,
		0,
		map[string]RawDataIndex{},
		0,
		0,
		0,
		0,
		0,
	}

	firstSegment := readTDMSSegment(file, 0, 1, emptySegment)
	readTDMSSegment(file, int64(firstSegment.nextSegPos), 0, firstSegment)

	finalPos, err := file.Seek(0, 1)

	if err != nil {
		log.Fatal(err)
	}

	file.Close()
	log.Printf("TDMS Closed at position: %d\n", finalPos)
}

func initLogging() {
	// If the file doesnt exit create it, or append to the file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Error return from os.OpenFile: ", err)
	}

	// Creates a MultiWrite, duplicates its write sto all the provided writers
	// In this instance, the file and stdout
	mw := io.MultiWriter(os.Stdout, file)

	log.SetOutput(mw)

	log.Println("TDMS Reader Init")
}

func displayTDMSGroupChannels(file *os.File, offset int64, whence int) {
	// Keep Log of Segments
	// Start at Segment 0
	// Seek to Begining of File
	// Until EOF
	//	read segment meta data:
	//		read lead in to vars
	//		read segment to vars
	//		read properties to vars
	//			take in file, previous segment data, previous segment?, whether hte next segment offset was set
	//			check ToCMask
	//				reuse/endianess/objlist
	//	update object metadata w segment?
	//  update object properties w properties?
	//  append segment
	//  previous segment = segment
	//  segment pos = segment.next_segment_pos
	//  seek file to next segment pos
	//
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
	log.Printf("Reading TDMS Segement starting at: %d", startPos)

	// Read TDMS Lead In
	// leadIn := readTDMSLeadIn(file, offset, whence)
	leadIn := readTDMSLeadIn(file, 0, 1)

	// Read TDMS Meta Data
	objMap := readTDMSMetaData(file, 0, 1, leadIn, prevSegment)

	// Calculate Number of Chunks
	numChunks := calculateChunks(objMap, leadIn.nextSegPos, leadIn.dataPos)

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
	// 		log.Printf("Read %s Values\n", key)
	// 		log.Printf("Number of Values: %d\n", len(data))
	// 		log.Printf("Max Value: %.6f\n", dataMax)
	// 		log.Printf("Min Value: %.6f\n", dataMin)
	// 		log.Printf("Average Value: %.6f\n", averageFloat64Slice(data))
	// 	}
	// }
	return Segment{
		uint64(startPos),
		numChunks,
		objMap,
		leadIn.ToCMask,
		leadIn.nextSegPos,
		leadIn.dataPos,
		0, //TODO:
		0, //TODO:
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

	log.Println("READING LEAD-IN")

	// Starts with a 4-byte tag that identifies a TDMS Segment ("TDSm")
	segStartTag := make([]byte, 4)
	_, err = io.ReadFull(file, segStartTag)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in readTDMSLeadIn: ", err)
	}
	if string(segStartTag) != "TDSm" {
		log.Fatal("Segment is not a TDMS")
	}
	log.Println("Valid TDMS Segment Starting at: ", segmentStartPos)

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
	log.Println("ToC BitMask: ", tocBitMask)
	if (kTocMetaData & tocBitMask) == kTocMetaData {
		log.Println("Segment Contains Meta Data")
	}
	if (0b1000 & tocBitMask) == 0b1000 {
		log.Println("Segment Contains Raw Data")
	}
	if (0b10000000 & tocBitMask) == 0b10000000 {
		log.Println("Segment Contains DAQmx Raw Data")
	}
	if (0b100000 & tocBitMask) == 0b100000 {
		log.Println("Segment Contains Interleaved Data")
	}
	if (0b1000000 & tocBitMask) == 0b1000000 {
		log.Println("Segment Contains Big Endian Data")
	}
	if (0b100 & tocBitMask) == 0b100 {
		log.Println("Segment Contains New Object List")
	}

	// 4 Byte Version Number
	// 4713 = v2.0
	// 4712 = Older
	versionNumber := uint32FromTDMS(file, 0, 1)
	log.Println("Version Number: ", versionNumber)

	// 8 Bytes - Length of Remaining Segment
	// Also known as Next Segment Offset
	// Remaining Length = Overall Length of Segment - Length of Lead in ()
	// If an application encounters a problem writing, all bytes will = 0xFF
	// can only happen at EOF
	segLength := uint64FromTDMS(file, 0, 1)
	log.Println("Segment Length: ", segLength)

	// 8 Bytes - Length of Metadata in Segment
	// Also known as raw data offset
	// If segment contains no metadata will = 0
	metaLength := uint64FromTDMS(file, 0, 1)
	log.Println("Metadata Length: ", metaLength)

	leadInSize := uint64(28)

	nextSegPos := uint64(0)
	if segLength == 0xFFFFFFFFFFFFFFFF {
		log.Printf("Segment incomplete, attempting to Read")
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
func readTDMSMetaData(file *os.File, offset int64, whence int, leadin LeadInData, prevSegment Segment) map[string]RawDataIndex {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// True if no MetaData
	if (kTocMetaData & leadin.ToCMask) != kTocMetaData {
		log.Println("Reuse Previous Segment Metadata")
		// TODO: Return + Implement
	}

	// TODO: Big Endianness with TocMask

	// TODO: if newObjList or NOT Preiouvs Segment????
	//			 existing Objects = None
	//			 orderedObjects = []
	//			 else copy objects from previous

	prevSegObjectNum := len(prevSegment.objects)

	// Initialize Empty Map for Objects
	objMap := make(map[string]RawDataIndex)

	if ((kTocNewObjList & leadin.ToCMask) == kTocNewObjList) || prevSegObjectNum == 0 {
	} else {
		// There can be a list of new objects that are appended,
		// or previous objects that are repeated with changed properties
		objMap = prevSegment.objects
	}

	log.Println("READING METADATA")

	// First 4 Bytes have number of objects in metadata
	numObjects := uint32FromTDMS(file, 0, 1)
	log.Println("Number of Objects: ", numObjects)

	// ar objects = make([]string, numObjects)
	for i := uint32(0); i < numObjects; i++ {
		log.Printf("Reading Object %d \n", i)

		// Read Object Path
		objPath := stringFromTDMS(file, 0, 1)
		log.Printf("Object %d Path: %s\n", i, objPath)

		// TODO: FINISH
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
		log.Printf("Object Raw Data Index: % x", rawDataIndexHeaderBytes)

		/*
			First Check if Already Have Object in List from Previous Segment
			if Object Already Exists:															DONE?
				Update the existing Obejct??												DONE?
			else if object path was in previous segment objects:	DONE?
				reuse previous segment objects											DONE?
			else:
				if raw data index matches previous:
					raise error, haven't seen object before
				else if rawDataPresent:															DONE
					read raw data index																DONE
			read num properties																		DONE
			read properties																				DONE
			calcaulte chunks
		*/

		// check if we already have objPath
		// only executes if present true, present will be true if found in map
		// first value is the value found
		if _, present := objMap[objPath]; present {
			// if true, update Object
			// TODO: Check
			objMap[objPath] = RawDataIndex{}
		} else if val, present := prevSegment.objects[objPath]; present {
			// reuse previous object
			objMap[objPath] = val
		} else {

		}

		noRawDataPresent := bytes.Compare(rawDataIndexHeaderBytes, noRawDataValue)

		// no Raw Data is Present
		if noRawDataPresent == 0 {
		} else {
			rawDataIndex := readTDMSRawDataIndex(file, 0, 1, rawDataIndexHeaderBytes)
			objMap[objPath] = rawDataIndex
		}

		// Number of Object Properties
		numProperties := uint32FromTDMS(file, 0, 1)
		log.Printf("Number of Object %d Properties: %d\n", i, numProperties)

		// Read Properties
		for j := uint32(0); j < numProperties; j++ {
			log.Printf("Reading Object %d Property %d\n", i, j)
			readTDMSProperty(file, 0, 1)

		}
	}
	return objMap
}

// Read the Properties for a TDMS Object
// TODO
// Change this to output a list of properties
func readTDMSProperty(file *os.File, offset int64, whence int) {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Property Name
	propertyName := stringFromTDMS(file, 0, 1)
	log.Printf("Property Name: %s\n", propertyName)

	// Printed in Hex
	propertyDataType := uint32FromTDMS(file, 0, 1)
	propertyTdsDataType := tdsDataType(propertyDataType)
	// TODO
	// Finish This
	switch propertyTdsDataType {
	default:
		log.Fatal("Property Data Type Unkown")
	case String:
		stringValue := stringFromTDMS(file, 0, 1)
		log.Printf("Property Value: %s\n", stringValue)
	case Int32:
		int32Value := int32FromTDMS(file, 0, 1)
		log.Printf("Property Value: %d\n", int32Value)
	case Uint32:
		uint32Value := uint32FromTDMS(file, 0, 1)
		log.Printf("Property Value: %d\n", uint32Value)
	case Uint64:
		uint64Value := uint64FromTDMS(file, 0, 1)
		log.Printf("Property Value: %d\n", uint64Value)
	case DBL:
		DBLValue := DBLFromTDMS(file, 0, 1)
		log.Printf("Property Value: %.6f\n", DBLValue)
	case Timestamp:
		log.Printf("Timestamp Property In Testing\n")
		timestampValue := timeFromTDMS(file, 0, 1)
		log.Printf("Property Value: %s\n", timestampValue.String())
	}
}

func readTDMSRawDataIndex(file *os.File, offset int64, whence int, rawDataIndexHeader []byte) RawDataIndex {
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return by file.Seek in readTDMSRawDataIndex: ", err)
	}

	indexLength := binary.LittleEndian.Uint32(rawDataIndexHeader)
	log.Printf("Object Index Length: %d\n", indexLength)

	dataType := tdsDataType(uint32FromTDMS(file, 0, 1))
	log.Printf("Object Data Type: %d\n", dataType)

	// must equal 1 for v2.0
	arrayDimension := uint32FromTDMS(file, 0, 1)
	if arrayDimension != 1 {
		log.Fatal("Not Valid TDMS 2.0, Data Dimension is not 1")
	}

	numValues := uint64FromTDMS(file, 0, 1)
	log.Printf("Object Number of Values: %d\n", numValues)

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
	log.Printf("Channel Raw Data Size: %d\n", channelRawDataSize)

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
func calculateChunks(objects map[string]RawDataIndex, nextSegPos uint64, dataPos uint64) uint64 {
	dataSize := uint64(0)
	for _, e := range objects {
		dataSize += e.rawDataSize
	}
	log.Printf("Data Size: %d", dataSize)

	totalDataSize := nextSegPos - dataPos
	log.Printf("Total Data Size: %d", totalDataSize)

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
		// log.Printf("Value %d: %d\n", i, val)
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
		// log.Printf("Value %d: %d\n", i, val)
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
		// log.Printf("Value %d: %.2f\n", i, val)
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
