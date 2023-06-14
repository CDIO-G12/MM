package utils

import "math"

func Cos(angle int) float64 {
	return math.Cos(float64(angle))
}

func Sin(angle int) float64 {
	return math.Sin(float64(angle))
}

func Degrees(angle float64) int {
	return int(math.Round(angle * 180 / math.Pi))
}

func Atan(angle float64) float64 {
	return math.Atan(angle)
}
