package tdms

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Reads a string from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns String
func readString(file *os.File, offset int64, whence int) string {
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

// Reads an int32 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns int32
func readInt32(file *os.File, offset int64, whence int) int32 {
	value := readUint32(file, offset, whence)
	return int32(value)
}

// Reads a single uint32 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns uint32
func readUint32(file *os.File, offset int64, whence int) uint32 {
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

// Reads a []uint32 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns []uint32
func readUint32Array(file *os.File, number int64, offset int64, whence int) []uint32 {
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

// Reads an int64 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns int64
func readInt64(file *os.File, offset int64, whence int) int64 {
	value := readUint64(file, offset, whence)
	return int64(value)
}

// Reads a uint64 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns uint64
func readUint64(file *os.File, offset int64, whence int) uint64 {
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

// Reads a []uint64 from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns []uint64
func readUint64Array(file *os.File, number int64, offset int64, whence int) []uint64 {
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

// Reads a SGL from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns Float32
func readSGL(file *os.File, offset int64, whence int) float32 {
	value := readUint32(file, offset, whence)
	return math.Float32frombits(value)
}

// Reads a Slice of SGLS from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns []Float32
func readSGLArray(file *os.File, number int64, offset int64, whence int) []float32 {
	size := int64(4)

	_, err := file.Seek(offset, whence)
	if err != nil {
		log.Fatal("Error return from file.Seek in DBLArrayFromTDMS: ", err)
	}

	intByteArray := make([]byte, number*size)
	_, err = io.ReadFull(file, intByteArray)
	if err != nil {
		log.Fatal("Error return from io.ReadFull in DBLArrayFromTDMS: ", err)
	}

	var vals []float32

	for i := int64(0); i < number; i++ {
		startBit := i * size
		endBit := startBit + size
		val := math.Float32frombits(binary.LittleEndian.Uint32(intByteArray[startBit:endBit]))
		vals = append(vals, val)
		// log.Debugf("Value %d: %.2f\n", i, val)
	}

	return vals
}

// Reads a DBL from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns Float64
func readDBL(file *os.File, offset int64, whence int) float64 {
	value := readUint64(file, offset, whence)
	return math.Float64frombits(value)
}

// Reads a []DBL from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns []Float64
func readDBLArray(file *os.File, number int64, offset int64, whence int) []float64 {
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

// Reads a Timestamp from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns time.Time
func readTime(file *os.File, offset int64, whence int) time.Time {
	posFractions := readUint64(file, offset, whence)
	LVseconds := readInt64(file, 0, 1)
	nanoSeconds := float64(posFractions) * math.Pow(2, -64) * 1e9
	secondsToUnix := 2.083e9
	timeValue := time.Unix(0, 0)
	if LVseconds != 0 && nanoSeconds != 0 {
		timeValue = time.Unix(LVseconds-int64(secondsToUnix), int64(nanoSeconds))
	}
	return timeValue
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

func getGroupsFromPathArray(paths []string) []string {
	var groups []string
	for _, path := range paths {
		if (path != "/") && (len(strings.Split(path, "/")) == 2) {
			groups = append(groups, path)
		}
	}
	return groups
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
