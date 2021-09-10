package main

import (
	"log"
	"os"
	"io"
	"encoding/hex"
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
	// kTocMetaData (1L << 1)
	// kTocRawData (1L << 3)
	// kTocDAQmxRawData (1L << 7)
	// kToxInterleavedData (1L << 5)
	// kTocBigEndian (1L << 6)
	// kTocNewObjList (1L << 2)
	tocBitMask := make([]byte, 4)
	_, err = io.ReadFull(file, tocBitMask)
	log.Println("NEED TO FINISH TOC MASK")
	log.Println("ToC BitMask string: ", hex.EncodeToString(tocBitMask))

	// 4 Byte Version Number
	// 4713 = v2.0
	// 4712 = Older
	versionNumberBytes := make([]byte, 4)
	_, err = io.ReadFull(file, versionNumberBytes)
	versionNumber := binary.LittleEndian.Uint32(versionNumberBytes)
	log.Println("Version Number Bytes:", versionNumberBytes)
	log.Println("Version Number String", hex.EncodeToString(versionNumberBytes))
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
	// var segment_position = 0

	// Read Segment Metadata
	// First 4 Bytes have number of objects in metadata
	numObjectsBytes := make([]byte, 4)
	_, err = io.ReadFull(file, numObjectsBytes)
	numObjects := binary.LittleEndian.Uint32(numObjectsBytes)
	log.Println("Number of Objects Bytes: ", numObjectsBytes)
	log.Println("Number of Objects String: ", hex.EncodeToString(numObjectsBytes))
	log.Println("Number of Objects: ", numObjects)

	// Check for Object?

	// Length of First Object Path
	firstObjPathLengthBytes := make([]byte, 4)
	_, err = io.ReadFull(file, firstObjPathLengthBytes)
	firstObjPathLength := binary.LittleEndian.Uint32(firstObjPathLengthBytes)
	log.Println("First Object Path Length Bytes: ", firstObjPathLengthBytes)
	log.Println("First Object Path Length String: ", hex.EncodeToString(firstObjPathLengthBytes))
	log.Println("First Object Path Length", firstObjPathLength)


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

