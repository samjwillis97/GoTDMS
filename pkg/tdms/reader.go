package tdms

import (
	log "github.com/sirupsen/logrus"
)

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

type RawDataIndex struct {
	dataType       tdsDataType
	arrayDimension uint32
	numValues      uint64
	rawDataSize    uint64
}

type SegmentObject struct {
	rawDataIndexHeader []byte
	rawDataIndex       RawDataIndex
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
