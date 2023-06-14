package utils

import "math"

func DegreeAdd(initial, next int) int {

	return CircularChecker(initial + next)
}

func DegreeSub(initial, next int) int {

	diff := math.Abs(float64(initial) - float64(next))

	if diff > 180 {
		return int(360 - diff)
	}

	return int(diff)
}

func CircularChecker(in int) (out int) {
	out = in
	if out > 359 {
		out -= 360
	}
	if out < 0 {
		out += 360
	}
	return
}
