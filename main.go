package main

import (
	"log"
	"os"
	"io"
	"bytes"
	"encoding/binary"
)


func main() {
	initLogging()

	file, err := os.OpenFile("test.tdms", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// Reading a TDMS File
	// https://www.ni.com/en-au/support/documentation/supplemental/07/tdms-file-format-internal-structure.html

	// Lead In
	startPosition, err := file.Seek(0, 1)

	// Starts with a 4-byte tag that identifies a TDMS Segment ("TDSm")
	segStartTag :=make([]byte, 4)
	_, err = io.ReadFull(file, segStartTag)
	if string(segStartTag) != "TDSm" {
		log.Fatal("Segment is not a TDMS")
	}
	log.Println("Valid TDMS Segment Starting at: ", startPosition)

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


	// MetaData

	// Read Segment Metadata
	// First 4 Bytes have number of objects in metadata
	numObjectsBytes := make([]byte, 4)
	_, err = io.ReadFull(file, numObjectsBytes)
	numObjects := binary.LittleEndian.Uint32(numObjectsBytes)
	log.Println("Number of Objects: ", numObjects)

	// Length of First Object Path
	firstObjPathLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(file, firstObjPathLengthBytes)
	firstObjPathLength := binary.LittleEndian.Uint32(firstObjPathLengthBytes)
	log.Println("First Object Path Length: ", firstObjPathLength)

	// Read Object Path
	firstObjPathBytes := make([]byte, firstObjPathLength)
	_, err = io.ReadFull(file, firstObjPathBytes)
	log.Println("First Object Path: ", string(firstObjPathBytes))

	// Object
	// Raw Data Index
	// FF FF FF FF means there is no raw data
	rawDataIndexBytes := make([]byte, 4)
	_,err = io.ReadFull(file, rawDataIndexBytes)
	log.Println("Raw Data Index: ", rawDataIndexBytes)
	noRawDataValue := []byte{255, 255, 255, 255}
	rawDataPresent := bytes.Compare(rawDataIndexBytes, noRawDataValue)
	if rawDataPresent == 0 {
		log.Println("No Raw Data Present")
	} else {
		log.Println("Raw Data Present")
	}

	// Group Properties
	numGroupOnePropertiesBytes := make([]byte, 4)
	_, err = io.ReadFull(file, numGroupOnePropertiesBytes)
	numGroupOneProperties := binary.LittleEndian.Uint32(numGroupOnePropertiesBytes)
	log.Println("Number of Group Properties: ", numGroupOneProperties)

	// Length of First Property Name
	firstPropertyNameLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(file, firstPropertyNameLengthBytes)
	firstPropertyNameLength := binary.LittleEndian.Uint32(firstPropertyNameLengthBytes)
	log.Println("First Property Name Length: ", firstPropertyNameLength)

	// First Property Name
	firstPropertyNameBytes := make([]byte, firstPropertyNameLength)
	_, err = io.ReadFull(file, firstPropertyNameBytes)
	log.Println("First Property Name: ", string(firstPropertyNameBytes))

	// First Property DataType
	firstPropertyDataTypeBytes := make([]byte, 4)
	_, err = io.ReadFull(file, firstPropertyDataTypeBytes)
	log.Println("First Property Data Type: ", firstPropertyDataTypeBytes)


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

