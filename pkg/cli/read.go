package cli

import (
	"fmt"
	"math"
	"os"
	"text/tabwriter"

	"github.com/samjwillis97/GoTDMS/pkg/analysis"
	"github.com/samjwillis97/GoTDMS/pkg/tdms"
	log "github.com/sirupsen/logrus"
)

func DisplayChannelRawData(file *os.File, channelPath string, length int64, offset uint64, allSegments []tdms.Segment, allProps map[string]map[string]tdms.Property) {
	// Determine Data Type of Segment
	// if TWF, defined by the properties
	// return RMS, P-P, CF for the whole file, add option for Block-by-block, that returns a slice
	firstSeg := true

	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	// Iterate through all File Segments
	for i, segment := range allSegments {
		// Iterate through all the objects in order
		// Skipping over the data we don't need to read
		for _, objPath := range segment.ObjectOrder {
			obj := segment.Objects[objPath]

			if objPath == channelPath {

				_, wfStartPresent := allProps[objPath]["wf_start_time"]
				_, wfStartOffsetPresent := allProps[objPath]["wf_start_offset"]
				_, wfIncrementPresent := allProps[objPath]["wf_increment"]
				_, wfSamplesPresent := allProps[objPath]["wf_samples"]

				// fmt.Println("New Segment:", i)
				// fmt.Println(allProps[objPath])

				if wfStartPresent && wfStartOffsetPresent && wfIncrementPresent && wfSamplesPresent {
					log.Debugln("Waveform Present")

					wf_increment := tdms.ReadDBL(file, allProps[objPath]["wf_increment"].ValuePosition, 0)
					wf_samples := tdms.ReadInt32(file, allProps[objPath]["wf_samples"].ValuePosition, 0)
					wf_start_time := tdms.ReadTime(file, allProps[objPath]["wf_start_time"].ValuePosition, 0)

					if firstSeg {
						fmt.Printf("TDMS Path:\t%s\n", channelPath)
						fmt.Printf("Sample Rate:\t%d Hz\n", int(1/wf_increment))
						fmt.Printf("Segment Length Length:\t%d Samples\n", wf_samples)
						fmt.Printf("Start Time: \t%s\n", wf_start_time)
						fmt.Printf("Total Segments:\t%d\n", len(allSegments))

						fmt.Fprintf(writer, "\nSeg No. \tRMS \tP-P \tCF\n")
					}

					_, err := file.Seek(int64(segment.DataPos), 0)
					if err != nil {
						log.Fatalln("Error from file.Seek in readChannelRawData")
					}

					data := make([]float64, 0)

					switch obj.RawDataIndex.DataType {
					case tdms.SGL:
						dataSGL := tdms.ReadSGLArray(file, int64(obj.RawDataIndex.NumValues), int64(obj.RawDataIndex.RawDataSize), 1)

						//convert data to float64
						for _, val := range dataSGL {
							data = append(data, float64(val))
						}

					case tdms.DBL:
						data = tdms.ReadDBLArray(file, int64(obj.RawDataIndex.NumValues), int64(obj.RawDataIndex.RawDataSize), 1)

					default:
						log.Fatal("Data Type Not Implemented")
					}

					rms := analysis.RmsFloat64Slice(data)
					min, max := analysis.MinMaxFloat64Slice(data)
					pp := math.Abs(max - min)
					cf := max / rms

					// fft, _ := analysis.VibFFT(data, wf_increment, 0)

					// fmt.Println(analysis.MaxFloat64(fft))
					// fmt.Println()

					fmt.Fprintf(writer, "%d \t%.4f \t%.4f \t%.4f\n", i, rms, pp, cf)

					firstSeg = false
				}
			}
		}
	}
	writer.Flush()
}
