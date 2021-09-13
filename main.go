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

type leadInData struct {
	ToCMask       uint32
	versionNumber uint32
	nextSegOffset uint64
	rawDataOffset uint64
}

type rawDataInfo struct {
	rawDataIndex   []byte
	dataType       tdsDataType
	arrayDimension uint32
	numValues      uint64
	rawDataSize    uint64
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

func main() {
	initLogging()

	file, err := os.OpenFile("testFiles/demo.tdms", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal("Error return from os.OpenFile: ", err)
	}

	// Reading a TDMS File
	// https://www.ni.com/en-au/support/documentation/supplemental/07/tdms-file-format-internal-structure.html
	readTDMSSegment(file, 0, 1)

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

// Reads a TDMS Segment
// Includes:
// - Lead In
// - Meta Data
// Data is written in Segments, every time data is appended to a TDMS, a new segment is created
// A segment consists of Lead In, Meta Data, and Raw Data.
// There are exceptions to the rules
// hence Different Groups when written after each other will be in different seg
func readTDMSSegment(file *os.File, offset int64, whence int) {
	readTDMSLeadIn(file, offset, whence)
	readTDMSMetaData(file, 0, 1)
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
func readTDMSLeadIn(file *os.File, offset int64, whence int) leadInData {
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
	// kTocMetaData (1L << 1)					= 0b0000 0010			= 0x2
	// kTocRawData (1L << 3)					= 0b0000 1000			= 0x8
	// kTocDAQmxRawData (1L << 7)			= 0b1000 0000			= 0x80
	// kToxInterleavedData (1L << 5)  = 0b0010 0000			= 0x20
	// kTocBigEndian (1L << 6)				= 0b0100 0000			= 0x40
	// kTocNewObjList (1L << 2)				= 0b0000 0100			= 0x4
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
	if (0b10 & tocBitMask) == 0b10 {
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

	return leadInData{
		tocBitMask,
		versionNumber,
		segLength,
		metaLength,
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
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func readTDMSMetaData(file *os.File, offset int64, whence int) {
	log.Println("READING METADATA")

	// First 4 Bytes have number of objects in metadata
	numObjects := uint32FromTDMS(file, 0, 1)
	log.Println("Number of Objects: ", numObjects)

	// var objects = make([]string, numObjects)
	for i := uint32(0); i < numObjects; i++ {
		log.Printf("Reading Object %d \n", i)

		// Read Object Path
		objPath := stringFromTDMS(file, 0, 1)
		log.Printf("Object %d Path: %s\n", i, objPath)

		// Read Object
		readTDMSObjectInfo(file, 0, 1)
		// objRawDataInfo := readTDMSObjectInfo(file, 0, 1)

		// Number of Object Properties
		numProperties := uint32FromTDMS(file, 0, 1)
		log.Printf("Number of Object %d Properties: %d\n", i, numProperties)

		for j := uint32(0); j < numProperties; j++ {
			log.Printf("Reading Object %d Property %d\n", i, j)
			readTDMSProperty(file, 0, 1)

		}
	}
	// TODO
	// Read the Data
}

// Read the Info for a TDMS Object
// Includes:
// - Raw Data Index
// - Raw Data Index Information
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func readTDMSObjectInfo(file *os.File, offset int64, whence int) rawDataInfo {
	// Check first 4 Bytes
	// if FF FF FF FF No Raw Data -> Read Properties
	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in readTDMSObject: ", err)
	}

	// Read Raw Data Index/Length of Index Information
	// FF FF FF FF means there is no raw data
	// 69 12 00 00 DAQmx Format Changing Scaler
	// 69 13 00 00 DAQmx Digital Line Scaler
	// Matches Previous Segment Same Object i.e. use previous
	// Otherwise
	rawDataIndexBytes := make([]byte, 4)
	_, err = io.ReadFull(file, rawDataIndexBytes)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in readTDMSObject: ", err)
	}
	indexLength := binary.LittleEndian.Uint32(rawDataIndexBytes)
	log.Printf("Object Raw Data Index: % x", rawDataIndexBytes)

	noRawDataValue := []byte{255, 255, 255, 255}
	// IMPLEMENT DAQMX

	rawDataPresent := bytes.Compare(rawDataIndexBytes, noRawDataValue)

	if rawDataPresent == 0 {
		log.Printf("Object No Raw Data Present\n")
		// to -> Read Properties
		return rawDataInfo{
			rawDataIndexBytes,
			0,
			0,
			0,
			0,
		}
	} else {
		// Raw Data is Present
		log.Printf("Object Index Length: %d\n", indexLength)

		// Get Index Information

		// tdsDataTypes from npTDMS
		// 0 = Void Pass
		// 1 = Int8, Size = 1
		// 2 = Int16, Size = 2
		// 3 = Int32, Size = 4
		// 4 = Int64, Size = 8
		// 5 = Uint8, Size = 1
		// 6 = Uint16, Size = 2
		// 7 = Uint32, Size = 4
		// 8 = Uint64, Size = 8
		// 9 = SGL, Size = 4
		// 10 = DBL, Size = 8
		// 11 = EXT Pass
		// 0x19 = SGL w Unit, Size = 4
		// 0x1A = DBL w Unit, Size = 8
		// 0x1B = EXT w Unit Pass
		// 0x20 = String
		// 0x21 = Boolean, Size = 1
		// 0x44 = Time, Size = 16
		// 0x08000C = Complex SGL
		// 0x10000d = Complex DBL
		// 0xFFFFFF = DAQmx Raw Data

		dataType := uint32FromTDMS(file, 0, 1)
		log.Printf("Object Data Type: %d\n", dataType)

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

		// must equal 1 for v2.0
		arrayDimension := uint32FromTDMS(file, 0, 1)
		if arrayDimension != 1 {
			log.Fatal("Not Valid TDMS 2.0, Data Dimension is not 1")
		}

		numValues := uint64FromTDMS(file, 0, 1)
		log.Printf("Object Number of Values: %d\n", numValues)

		// TODO
		// If String Read Value Size

		// to -> Read Properties

		channelRawDataSize := uint64(dataSize) * uint64(arrayDimension) * numValues
		log.Printf("Channel Raw Data Size: %d\n", channelRawDataSize)

		return rawDataInfo{
			rawDataIndexBytes,
			tdsDataType(dataType),
			arrayDimension,
			numValues,
			channelRawDataSize,
		}
	}
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
	LVseconds := int64FromTDMS(file, offset, whence)
	posFractions := uint64FromTDMS(file, 0, 1)
	nanoSeconds := float64(posFractions) * math.Pow(2, -64) * 1e9
	secondsToUnix := 2.083e9
	timeValue := time.Unix(0, 0)
	if LVseconds != 0 && nanoSeconds != 0 {
		timeValue = time.Unix(LVseconds-int64(secondsToUnix), int64(nanoSeconds))
	}
	return timeValue
}
