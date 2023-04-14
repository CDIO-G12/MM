package main

import "math"

type categoryType int

const (
	NA categoryType = iota
	ball
	robot
	emergency
	goal
	corner
	middleXcorner
)

type pointType struct {
	x     int
	y     int
	angle int
}

type poiType struct {
	point    pointType
	category categoryType
}

func (p1 pointType) dist(p2 pointType) (angle int, len int) {
	first := math.Pow(float64(p2.x-p1.x), 2)
	second := math.Pow(float64(p2.y-p1.y), 2)
	len = int(math.Sqrt(first + second))
	dx := p1.x - p2.x
	dy := p1.y - p2.y
	theta := math.Atan2(float64(dy), float64(dx))
	theta *= 180 / math.Pi
	angle = int(theta) //+ 180

	return
}
