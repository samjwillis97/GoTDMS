package main

import (
	"log"
	"os"
	"io"
	"bytes"
	"encoding/binary"
)

type leadInData struct {
	ToCMask uint32
	versionNumber uint32
	nextSegOffset uint64
	rawDataOffset uint64
}

func main() {
	initLogging()

	file, err := os.OpenFile("test.tdms", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Reading a TDMS File
	// https://www.ni.com/en-au/support/documentation/supplemental/07/tdms-file-format-internal-structure.html

	// initLeadInData := readTDMSLeadIn(file, 0, 0)
	readTDMSLeadIn(file, 0 , 0)
	readTDMSMetaData(file, 0, 1)

	if err != nil {
		log.Fatal(err)
	}

	file.Close()
}

func initLogging() {
	// If the file doesnt exit create it, or append to the file
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil { log.Fatal(err) }
	
	// Creates a MultiWrite, duplicates its write sto all the provided writers
	// In this instance, the file and stdout
	mw := io.MultiWriter(os.Stdout, file)

	log.SetOutput(mw)

	log.Println("TDMS Reader Init")
}

// Reads the TDMS Lead-In
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
func readTDMSLeadIn(file *os.File, offset int64, whence int) leadInData {
	segmentStartPos, err := file.Seek(offset, whence)

	log.Println("READING LEAD-IN")

	// Starts with a 4-byte tag that identifies a TDMS Segment ("TDSm")
	segStartTag :=make([]byte, 4)
	_, err = io.ReadFull(file, segStartTag)
	if string(segStartTag) != "TDSm" {
		log.Fatal("Segment is not a TDMS")
	}
	log.Println("Valid TDMS Segment Starting at: ", segmentStartPos)

	// 4 Byte ToC BitMask NEED TO WORK OUT
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
	tocBitMask := binary.LittleEndian.Uint32(tocBitMaskBytes)	
	log.Println("ToC BitMask: ", tocBitMask)
	if ((0b10 & tocBitMask) == 0b10) {
		log.Println("Segment Contains Meta Data")	
	}
	if ((0b1000 & tocBitMask) == 0b1000) {
		log.Println("Segment Contains Raw Data")	
	}
	if ((0b10000000 & tocBitMask) == 0b10000000) {
		log.Println("Segment Contains DAQmx Raw Data")	
	}
	if ((0b100000 & tocBitMask) == 0b100000) {
		log.Println("Segment Contains Interleaved Data")	
	}
	if ((0b1000000 & tocBitMask) == 0b1000000) {
		log.Println("Segment Contains Big Endian Data")	
	}
	if ((0b100 & tocBitMask) == 0b100) {
		log.Println("Segment Contains New Object List")	
	}  

	// 4 Byte Version Number
	// 4713 = v2.0
	// 4712 = Older
	versionNumberBytes := make([]byte, 4)
	_, err = io.ReadFull(file, versionNumberBytes)
	versionNumber := binary.LittleEndian.Uint32(versionNumberBytes)
	log.Println("Version Number: ", versionNumber)

	// 8 Bytes - Length of Remaining Segment
	// Also known as Next Segment Offset
	// Remaining Length = Overall Length of Segment - Length of Lead in ()
	// If an application encounters a problem writing, all bytes will = 0xFF
	// can only happen at EOF
	segLengthBytes := make([]byte, 8)
	_, err = io.ReadFull(file, segLengthBytes)
	segLength := binary.LittleEndian.Uint64(segLengthBytes)
	log.Println("Segment Length: ", segLength)

	// 8 Bytes - Length of Metadata in Segment
	// Also known as raw data offset
	// If segment contains no metadata will = 0
	metaLengthBytes := make([]byte, 8)
	_, err = io.ReadFull(file, metaLengthBytes)
	metaLength := binary.LittleEndian.Uint64(metaLengthBytes)
	log.Println("Metadata Length: ", metaLength)

	if err != nil {
		log.Fatal(err)
	}

	return leadInData{
		tocBitMask,
		versionNumber,
		segLength,
		metaLength,
	}
}

// Add Check for Raw Data Index
func readTDMSMetaData(file *os.File, offset int64, whence int) {
	log.Println("READING METADATA")

	// First 4 Bytes have number of objects in metadata
	numObjectsBytes := make([]byte, 4)
	_, err := io.ReadFull(file, numObjectsBytes)
	numObjects := binary.LittleEndian.Uint32(numObjectsBytes)
	log.Println("Number of Objects: ", numObjects)

	// var objects = make([]string, numObjects)

	for i := uint32(0); i < numObjects; i++ {
		// Length of Object Path
		objPathLengthBytes := make([]byte, 4)
		_, err = io.ReadFull(file, objPathLengthBytes)
		objPathLength := binary.LittleEndian.Uint32(objPathLengthBytes)

		// Read Object Path
		objPathBytes := make([]byte, objPathLength)
		_, err = io.ReadFull(file, objPathBytes)
		log.Printf("Object %d Path: %s\n", i, string(objPathBytes))
		// objects[i] = string(objPathBytes)

		// Read Object Raw Data Index
		// FF FF FF FF means there is no raw data
		rawDataIndexBytes := make([]byte, 4)
		_,err = io.ReadFull(file, rawDataIndexBytes)
		noRawDataValue := []byte{255, 255, 255, 255}
		rawDataPresent := bytes.Compare(rawDataIndexBytes, noRawDataValue)
		if rawDataPresent == 0 {
			log.Printf("Object %d No Raw Data Present\n", i)
		} else {
			log.Printf("Object %d Raw Data Present\n", i)
		}

		// Number of Object Properties
		numGroupPropertiesBytes := make([]byte, 4)
		_, err = io.ReadFull(file, numGroupPropertiesBytes)
		numGroupProperties := binary.LittleEndian.Uint32(numGroupPropertiesBytes)
		log.Printf("Number of Object %d Group Properties: %d\n", i, numGroupProperties)

		for j := uint32(0); j < numGroupProperties; j++ {
			// Length of Property Name
			propertyNameLengthBytes := make([]byte, 4)
			_, err = io.ReadFull(file, propertyNameLengthBytes)
			propertyNameLength := binary.LittleEndian.Uint32(propertyNameLengthBytes)

			// Property Name
			propertyNameBytes := make([]byte, propertyNameLength)
			_, err = io.ReadFull(file, propertyNameBytes)
			log.Printf("Object %d Group %d Property Name: %s\n", i, j, string(propertyNameBytes))

			// Property Data Type
			// tdsTypeString			0x20
			// tdsTypeBoolean			0x21
			// tdsTypeTimeStamp		0x44
			// Printed in Hex
			propertyDataTypeBytes := make([]byte, 4)
			_, err = io.ReadFull(file, propertyDataTypeBytes)
			propertyDataType := binary.LittleEndian.Uint32(propertyDataTypeBytes)
			switch propertyDataType {
			case 0x20:
				log.Printf("Object %d Group %d Property Data Type: String", i, j)

				// Length of String
				stringLengthBytes := make([]byte, 4)
				_, err = io.ReadFull(file, stringLengthBytes)
				stringLength := binary.LittleEndian.Uint32(stringLengthBytes)

				// String Value
				stringValueBytes := make([]byte, stringLength)
				_, err = io.ReadFull(file, stringValueBytes)
				log.Printf("Object %d Group %d Property Value: %s\n", i, j, string(stringValueBytes))
			}
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}
