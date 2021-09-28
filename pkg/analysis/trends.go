package analysis

import (
	"math"
)

// Finds Min and Max of a []float64
// Returns Minimum and Maximum Value
func MinMaxFloat64Slice(y []float64) (min float64, max float64) {
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

// Finds Root Mean Square of []float64
// Returns Value (float64)
func RmsFloat64Slice(y []float64) (rms float64) {
	meanSqr := math.Pow(y[0], 2)

	y = y[1:]

	for _, v := range y {
		meanSqr += math.Pow(v, 2)
	}

	return math.Sqrt(meanSqr / float64(len(y)+1))
}

func PpFloat64Slice(y []float64) (pp float64) {
	min,max := MinMaxFloat64Slice(y)
	return math.Abs(max - min)
}

// Finds Max Value of a []float64
// Returns Index in Slice (int), and Max Value (float64)
func MaxFloat64(y []float64) (ndx int, val float64) {
	max := y[0]
	index := 0

	for i, v := range y {
		if v > max {
			// fmt.Printf("%d %e\n", i, v)
			max = v
			index = i
		}
	}

	return index, max
}

// Finds average of a []float64
// Returns average(float64)
func AverageFloat64Slice(y []float64) (avg float64) {
	avg = y[0]

	y = y[1:]

	for _, v := range y {
		avg += v
	}

	return avg / float64(len(y)+1)
}
