package ahrs

import "math"

// AHRS algorithm interface
type AHRS interface {
	// Update9D updates position using 9D, returning quaternions
	Update9D(gx, gy, gz, ax, ay, az, mx, my, mz float64) [4]float64
	// Update6D updates position using 6D, returning quaternions
	Update6D(gx, gy, gz, ax, ay, az float64) [4]float64
}

func invSqrt(x float64) float64 {
	return 1 / math.Sqrt(x)
}
