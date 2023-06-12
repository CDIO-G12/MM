package utils

import (
	"math"
	"sync"
)

type CategoryType int

const (
	NA CategoryType = iota
	Ball
	Goal
	Corner
	MiddleXcorner
	WayPoint
	Emergency
	Start
	Found
	NotFound
)

func (s CategoryType) String() string {
	switch s {
	case NA:
		return "na"
	case Ball:
		return "ball"
	case Goal:
		return "goal"
	case Corner:
		return "corner"
	case MiddleXcorner:
		return "middle x corner"
	case WayPoint:
		return "waypoint"
	case Emergency:
		return "emergency"
	case Start:
		return "start"
	case Found:
		return "found"
	case NotFound:
		return "not found"
	}
	return "cat"
}

type PixelDistType struct {
	Definition float64
	MU         sync.RWMutex
}

var pixelDist = PixelDistType{Definition: 1}

type PointType struct {
	X     int
	Y     int
	Angle int
}

type PoiType struct {
	Point    PointType
	Category CategoryType
}

type SafePointType struct {
	Point PointType
	MU    sync.Mutex
}

var CurrentPos = SafePointType{}

func (p *SafePointType) Get() PointType {
	p.MU.Lock()
	defer p.MU.Unlock()
	return p.Point
}

func (p *SafePointType) Set(new PointType) {
	p.MU.Lock()
	defer p.MU.Unlock()
	p.Point = new
}

func GetPixelDist() float64 {
	pixelDist.MU.RLock()
	defer pixelDist.MU.RUnlock()
	return pixelDist.Definition
}
func SetPixelDist(in float64) {
	pixelDist.MU.RLock()
	defer pixelDist.MU.RUnlock()
	pixelDist.Definition = in / TrackingDistance
	//log.Infoln("Updated pixel def: ", pixelDist.Definition)
}

func (p1 PointType) Dist(p2 PointType) (angle int, dist int) {
	first := math.Pow(float64(p2.X-p1.X), 2)
	second := math.Pow(float64(p2.Y-p1.Y), 2)
	dist = int(math.Sqrt(first + second))
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	theta := math.Atan2(float64(dy), float64(dx))
	theta *= 180 / math.Pi
	angle = int(theta) //+ 180
	angle = CircularChecker(angle)

	return
}

// compares 2 points to see if they are close to each other
func (old PointType) IsClose(new PointType, threshold int) bool {
	_, len := old.Dist(new)
	return len < threshold
}

// sort balls purely based on length to closest
func (currentPos PointType) SortBalls(balls []PointType) (sortedBalls []PointType, err error) {
	origLength := len(balls)
	if origLength < 2 {
		sortedBalls = balls
		return
	}
	//fmt.Println(currentPos.findRoute(balls, []pointType{}))

	for i := 0; i < origLength; i++ {
		minDist := 99999
		minI := 0
		for j, v := range balls {
			_, len := currentPos.Dist(v)
			if len < minDist {
				minDist = len
				minI = j
			}
		}
		sortedBalls = append(sortedBalls, balls[minI])
		currentPos = balls[minI]
		balls = remove(balls, minI)
	}
	return
}

// remove an element from a slice
func remove(slice []PointType, s int) []PointType {
	return append(slice[:s], slice[s+1:]...)
}

func (point PointType) CalcNextPos(distance int) PointType {
	radian := float64(point.Angle) * math.Pi / 180.0
	newX := point.X - int(float64(distance)*math.Cos(radian)+0.5)
	newY := point.Y - int(float64(distance)*math.Sin(radian)+0.5)
	return PointType{X: newX, Y: newY, Angle: point.Angle}
}
