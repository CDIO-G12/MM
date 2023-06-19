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
	PreciseWayPoint
	Emergency
	Start
	Found
	NotFound
	Calibrate
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
	case PreciseWayPoint:
		return "precise waypoint"
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

var pixelDist = PixelDistType{Definition: 0.5}

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
	point PointType
	mu    sync.RWMutex
}

var CurrentPos = SafePointType{}

func (p *SafePointType) Get() PointType {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.point
}

func (p *SafePointType) Set(new PointType) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.point = new
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

func (p1 PointType) AngleAndDist(p2 PointType) (angle int, dist int) {
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

func (p1 PointType) Dist(p2 PointType) (dist int) {
	first := math.Pow(float64(p2.X-p1.X), 2)
	second := math.Pow(float64(p2.Y-p1.Y), 2)
	dist = int(math.Sqrt(first + second))

	return
}

// compares 2 points to see if they are close to each other
func (old PointType) IsClose(new PointType, threshold int) bool {
	return old.Dist(new) < threshold
}

// remove an element from a slice
func Remove(slice []PointType, s int) []PointType {
	return append(slice[:s], slice[s+1:]...)
}

func (point PointType) CalcNextPos(distance int) PointType {
	radian := float64(point.Angle) * math.Pi / 180.0
	newX := point.X - int(float64(distance)*math.Cos(radian)+0.5)
	newY := point.Y - int(float64(distance)*math.Sin(radian)+0.5)
	return PointType{X: newX, Y: newY, Angle: point.Angle}
}
