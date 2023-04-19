package utils

func DegreeAdd(initial, next int) int {

	return CircularChecker(initial + next)
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
