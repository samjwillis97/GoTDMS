package tdms

import (
	"encoding/binary"
	"io"
	"math"
	"os"
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
func StringFromTDMS(file *os.File, offset int64, whence int) string {
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
func Int32FromTDMS(file *os.File, offset int64, whence int) int32 {
	value := Uint32FromTDMS(file, offset, whence)
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
func Uint32FromTDMS(file *os.File, offset int64, whence int) uint32 {
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
func Uint32ArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []uint32 {
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
func Int64FromTDMS(file *os.File, offset int64, whence int) int64 {
	value := Uint64FromTDMS(file, offset, whence)
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
func Uint64FromTDMS(file *os.File, offset int64, whence int) uint64 {
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
func Uint64ArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []uint64 {
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
func SGLFromTDMS(file *os.File, offset int64, whence int) float32 {
	value := Uint32FromTDMS(file, offset, whence)
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
func SGLArrayFromTDMS(file *os.File, number int64, offset int64, whence int) []float32 {
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
func DBLFromTDMS(file *os.File, offset int64, whence int) float64 {
	value := Uint64FromTDMS(file, offset, whence)
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

// Reads a Timestamp from a TDMS File
//
// Starts at Byte Defined by Offset
// When is the reference point for the offset
// 0 = Beginning of File
// 1 = Current Position
// 2 = End of File
//
// Returns time.Time
func TimeFromTDMS(file *os.File, offset int64, whence int) time.Time {
	posFractions := Uint64FromTDMS(file, offset, whence)
	LVseconds := Int64FromTDMS(file, 0, 1)
	nanoSeconds := float64(posFractions) * math.Pow(2, -64) * 1e9
	secondsToUnix := 2.083e9
	timeValue := time.Unix(0, 0)
	if LVseconds != 0 && nanoSeconds != 0 {
		timeValue = time.Unix(LVseconds-int64(secondsToUnix), int64(nanoSeconds))
	}
	return timeValue
}
