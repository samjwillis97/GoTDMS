package analysis

import (
	"math/cmplx"

	"github.com/mjibson/go-dsp/fft"
	"gonum.org/v1/gonum/dsp/window"
)

type SpectrumInfo struct {
	fMax	float64
	binSize float64
}

// Performs an FFT on a []float64
// Uses a Hanning Window by Default
//
// Returns FFT as a []float64 and FFT Information
func VibFFT(y []float64, dt float64, averages int) ([]float64, SpectrumInfo) {
	var result = make([]float64, 0)
	var specInfo SpectrumInfo

	if averages > 0 {
		avgLen := len(y) / averages
		specInfo.binSize = 1 / (float64(avgLen) / (1 / dt))

		for i := 0; i < averages; i++ {
			cut := y[i*avgLen : (i+1)*avgLen]
			cut = window.Hann(cut)
			vibFft := fft.FFTReal(cut)

			for j := 0; j < len(cut); j++ {
				mag, _ := cmplx.Polar(vibFft[j])
				if i == 0 {
					result = append(result, mag)
				} else {
					result[i] += mag
				}
			}
		}
		for k := 0; k < len(result); k++ {
			result[k] = result[k] / float64(averages)
		}
	} else {
		y = window.Hann(y)

		specInfo.binSize = 1 / (float64(len(y)) / (1 / dt))

		vibFft := fft.FFTReal(y)

		for i := 0; i < len(y); i++ {
			mag, _ := cmplx.Polar(vibFft[i])
			result = append(result, mag)
		}
	}
	specInfo.fMax = specInfo.binSize * (1 / dt)

	return result, specInfo
}
