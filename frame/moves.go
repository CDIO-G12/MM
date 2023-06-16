package frame

import (
	u "MM/utils"
	"fmt"
	"math"
)

func (f *FrameType) CreateMoves(nextPos u.PoiType) (directions []u.PoiType) {
	currentPos := u.CurrentPos.Get()

	if nextPos.Category == u.Goal {
		if currentPos.IsClose(nextPos.Point, 5) {
			directions = append(directions, nextPos)
			return
		}

		directions = make([]u.PoiType, 3)
		directions[2] = u.PoiType{Point: u.PointType{X: nextPos.Point.X + int(u.MmToGoal*u.GetPixelDist()), Y: nextPos.Point.Y}, Category: u.WayPoint}
		directions[1] = u.PoiType{Point: u.PointType{X: nextPos.Point.X + int(u.MmToGoal*u.GetPixelDist())/2, Y: nextPos.Point.Y}, Category: u.WayPoint}
		directions[0] = nextPos
		return
	}

	lastAppended := nextPos
	directions = append(directions, lastAppended)

	if nextPos.Category == u.Ball {
		if nextPos.Point.Angle >= u.RatingCorner {
			corner := nextPos.Point.Angle - u.RatingCorner
			cat := u.PreciseWayPoint
			for i := 1; i <= 2; i++ {
				point := nextPos.Point
				offset := int(u.DistanceFromBallCorner * u.GetPixelDist() * 0.707106781)
				fmt.Println("Corner ", corner, point, offset)
				switch corner {
				case 0: //UpLeft
					point.X += offset * i
					point.Y += offset * i
				case 1: //UpRight
					point.X -= offset * i
					point.Y += offset * i
				case 2: //DownRight
					point.X -= offset * i
					point.Y -= offset * i
				case 3: //DownLeft
					point.X += offset * i
					point.Y -= offset * i
				}
				if i == 2 {
					cat = u.WayPoint
				}
				lastAppended = u.PoiType{Point: point, Category: cat}
				directions = append(directions, lastAppended)
			}
		} else if nextPos.Point.Angle >= u.RatingBorder {
			border := nextPos.Point.Angle - u.RatingBorder
			point := nextPos.Point
			offset := int(u.DistanceFromBall * u.GetPixelDist())
			fmt.Println("Border ", border, point, offset)
			switch border {
			case 0: //Up
				point.Y += offset + 10
			case 1: //Left
				point.X += offset + 10
			case 2: //Down
				point.Y -= offset + 10
			case 3: //Right
				point.X -= offset + 10
			}
			lastAppended = u.PoiType{Point: point, Category: u.PreciseWayPoint}
			directions = append(directions, lastAppended)
		}
	}

	directions = append(directions, f.calcWaypointsNew(currentPos, lastAppended.Point)...)

	/*first := f.findClosestGuidePosition(currentPos)
	last := f.findClosestGuidePosition(nextPos.Point)
	*/
	//directions = append(directions, f.CalculateWaypoint(nextPos)...)
	/*TODO: Make middle positions

	Check if middle x is in the way, and add more points
	*/

	f.createTestImg(directions, "moves", f.MiddleX[:])

	//directions = append(directions, nextPos)

	return
}

func (f *FrameType) CalculateWaypoint(nextPos u.PoiType) (WayPoints []u.PoiType) {

	currentPos := u.CurrentPos.Get()

	middleXPoint := f.MiddleXPoint()

	angleX, distX := currentPos.Dist(middleXPoint)
	angleB, distB := currentPos.Dist(nextPos.Point)

	safeDist := float64(100)

	fmt.Println(distX, distB, angleX, angleB)
	//Check if the next ball is on the other side of the middle x
	if distX > distB {
		return
	}

	//Check if the ball is clear of the middle x
	if ((angleX + 10) < angleB) && ((angleX - 10) > angleB) {
		return
	}

	under := 1
	if nextPos.Point.Y < middleXPoint.Y {
		under = -1
	}

	x := middleXPoint.X + int(safeDist*math.Cos(float64(angleX)))*under
	y := middleXPoint.Y + int(safeDist*math.Sin(float64(angleX)))*under

	WayPoints = append(WayPoints, u.PoiType{Point: u.PointType{X: x, Y: y}, Category: u.WayPoint})
	fmt.Println("Added waypoint - ", WayPoints)

	return
}

func (f *FrameType) calcWaypoint(nextPos u.PoiType) (WayPoints []u.PoiType) {

	currentPos := u.CurrentPos.Get()

	middleXPoint := f.MiddleXPoint()

	angleX, distX := currentPos.Dist(middleXPoint)
	angleB, distB := currentPos.Dist(nextPos.Point)

	fmt.Println(distX, distB, angleX, angleB)

	if distB < distX {
		return
	}

	if ((angleX + 10) < angleB) && ((angleX - 10) > angleB) {
		return
	}

	WayPoints = append(WayPoints, u.PoiType{Point: f.findClosestGuidePosition(currentPos), Category: u.WayPoint})
	WayPoints = append(WayPoints, u.PoiType{Point: f.findClosestGuidePosition(nextPos.Point), Category: u.WayPoint})

	return
}

func (f *FrameType) FindThreeClosestXPoints() (points []u.PointType) {

	currentPos := u.CurrentPos.Get()

	points = make([]u.PointType, 3)

	// Find three closest points of the middle x
	for i, point := range f.MiddleX {
		if i == 0 {
			points[i] = point
			continue
		}

		_, dist := currentPos.Dist(point)
		_, dist2 := currentPos.Dist(points[i-1])

		if dist < dist2 {
			points[i] = point
		} else {
			points[i] = points[i-1]
			points[i-1] = point
		}
	}

	return points

}

const hardDist = 75
const borderDist = 25
const cornerDist = 25

func (f *FrameType) RateBall(ball *u.PointType) {
	f.MU.RLock()
	defer f.MU.RUnlock()
	pd := u.GetPixelDist()

	_, dist := f.MiddleXPoint().Dist(*ball)
	if dist < int(cornerDist*pd) {
		ball.Angle = u.RatingMiddleX
		return
	}

	for i, corner := range f.Corners {
		_, dist := corner.Dist(*ball)
		if dist < cornerDist {
			ball.Angle = u.RatingCorner + i
			return
		}
	}

	up := u.Avg(f.Corners[0].Y, f.Corners[1].Y)
	left := u.Avg(f.Corners[0].X, f.Corners[3].X)
	down := u.Avg(f.Corners[2].Y, f.Corners[3].Y)
	right := u.Avg(f.Corners[1].X, f.Corners[2].X)

	bordersDist := []int{u.Abs(ball.Y - up), u.Abs(ball.X - left), u.Abs(ball.Y - down), u.Abs(ball.X - right)}
	min := 999999
	minI := 5
	for i, border := range bordersDist {
		if border < min {
			minI = i
			min = border
		}
	}

	if min < borderDist {
		ball.Angle = u.RatingBorder + minI
	} else if min < hardDist {
		ball.Angle = u.RatingHard
	}

	return
}

func (f *FrameType) WithinBorder(point u.PointType) bool {
	f.MU.RLock()
	for _, v := range f.Corners {
		if v.X == 0 || v.Y == 0 {
			f.MU.RUnlock()
			return true
		}
	}
	up := u.Avg(f.Corners[0].Y, f.Corners[1].Y)
	left := u.Avg(f.Corners[0].X, f.Corners[3].X)
	down := u.Avg(f.Corners[2].Y, f.Corners[3].Y)
	right := u.Avg(f.Corners[1].X, f.Corners[2].X)
	f.MU.RUnlock()

	if point.Y <= up {
		return false
	}
	if point.X <= left {
		return false
	}
	if point.Y >= down {
		return false
	}
	if point.X >= right {
		return false
	}
	return true
}

func (f *FrameType) calcWaypointsNew(current, next u.PointType) (export []u.PoiType) {
	waypoints := []u.PoiType{}

	fmt.Println("Current ", current, "Next ", next)

	for i := 0; i < 5; i++ {
		//fmt.Println("Current ", current, "Next ", next, "I", i)
		current = f.checkIntersect(current, next)
		if current.X == 0 {
			break
		}
		waypoints = append(waypoints, u.PoiType{Point: current, Category: u.WayPoint})
	}

	//reverse the slice
	l := len(waypoints)
	export = make([]u.PoiType, l)
	for i, v := range waypoints {
		export[l-i-1] = v
	}

	return
}

func checkWithin(point, start, stop u.PointType, threshold int) bool {

	var maxX, minX, maxY, minY int

	if start.X > stop.X {
		maxX = start.X
		minX = stop.X
	} else {
		maxX = stop.X
		minX = start.X
	}

	if start.Y > stop.Y {
		maxY = start.Y
		minY = stop.Y
	} else {
		maxY = stop.Y
		minY = start.Y
	}

	withinX := (point.X > minX+threshold && point.X < maxX-threshold)
	withinY := (point.Y > minY+threshold && point.Y < maxY-threshold)

	return withinX || withinY

}

// check if line between current and next intersects middlex guidecorners
func (f *FrameType) checkIntersect(current, next u.PointType) (waypoint u.PointType) {

	intersectLRPoint := CalculateIntersection(current, next, f.guideCorners[0], f.guideCorners[1])
	intersectUDPoint := CalculateIntersection(current, next, f.guideCorners[2], f.guideCorners[3])

	if intersectLRPoint.X == 0 && intersectUDPoint.X == 0 {
		return
	}

	intersectLR := checkWithin(intersectLRPoint, current, next, 5) && checkWithin(intersectLRPoint, f.guideCorners[0], f.guideCorners[1], 5)
	intersectUD := checkWithin(intersectUDPoint, current, next, 5) && checkWithin(intersectUDPoint, f.guideCorners[2], f.guideCorners[3], 5)

	fmt.Println("HER!!!!", current, next, intersectLRPoint, intersectUDPoint, intersectLR, intersectUD)

	//fmt.Println("Waypoint: ", current, "Ball ", next, "LR Intersect: ", intersectLR, "UD Intersect: ", intersectUD, "LR Point: ", intersectLRPoint, "UD Point: ", intersectUDPoint)
	//fmt.Println("middleX: ", f.MiddleXPoint())

	// no intersections
	if !intersectLR && !intersectUD {
		return
	}

	// only intersect LR
	if intersectLR && !intersectUD {
		_, distL := intersectLRPoint.Dist(f.guideCorners[0])
		_, distR := intersectLRPoint.Dist(f.guideCorners[1])

		if distL < distR {
			waypoint = f.guideCorners[0]
		} else {
			waypoint = f.guideCorners[1]
		}
		return
	}

	// only intersect UD
	if !intersectLR && intersectUD {
		_, distU := intersectUDPoint.Dist(f.guideCorners[2])
		_, distD := intersectUDPoint.Dist(f.guideCorners[3])

		if distU < distD {
			waypoint = f.guideCorners[2]
		} else {
			waypoint = f.guideCorners[3]
		}
		return
	}

	//intersects both - find closest

	//find the best LR and UD guidecorner
	gcLR := u.PointType{}
	gcUD := u.PointType{}

	_, distL := intersectLRPoint.Dist(u.PointType{X: f.guideCorners[0].X, Y: f.guideCorners[0].Y})
	_, distR := intersectLRPoint.Dist(u.PointType{X: f.guideCorners[1].X, Y: f.guideCorners[1].Y})
	if distL < distR {
		gcLR = f.guideCorners[0]
	} else {
		gcLR = f.guideCorners[1]
	}
	_, distU := intersectUDPoint.Dist(u.PointType{X: f.guideCorners[2].X, Y: f.guideCorners[2].Y})
	_, distD := intersectUDPoint.Dist(u.PointType{X: f.guideCorners[3].X, Y: f.guideCorners[3].Y})
	if distU < distD {
		gcUD = f.guideCorners[2]
	} else {
		gcUD = f.guideCorners[3]
	}

	//find the gc closest to robot
	_, distLR := gcLR.Dist(current)
	_, distUD := gcUD.Dist(current)
	fmt.Println("dists", gcLR, gcUD, distLR, distUD)
	if distUD < distLR {
		waypoint = gcUD
	} else {
		waypoint = gcLR
	}

	return
}

type precisePoint struct {
	X, Y float64
}

func toPrecise(point u.PointType) precisePoint {
	return precisePoint{X: float64(point.X), Y: float64(point.Y)}
}

func fromPrecise(point precisePoint) u.PointType {
	return u.PointType{X: int(point.X), Y: int(point.Y)}
}

func (point precisePoint) toPoint() u.PointType {
	return fromPrecise(point)
}

func CalculateIntersection(p1, p2, p3, p4 u.PointType) u.PointType {
	line1Start, line1End, line2Start, line2End := toPrecise(p1), toPrecise(p2), toPrecise(p3), toPrecise(p4)

	// Check if Line 1 is vertical
	if line1End.X-line1Start.X == 0 {
		// Check if Line 2 is also vertical
		if line2End.X-line2Start.X == 0 {
			return u.PointType{}
		}

		// Calculate the slope and y-intercept of Line 2
		m2 := (line2End.Y - line2Start.Y) / (line2End.X - line2Start.X)
		b2 := line2Start.Y - m2*line2Start.X

		// Calculate the intersection point
		x := line1Start.X
		y := m2*x + b2

		intersectionPoint := precisePoint{X: x, Y: y}
		return intersectionPoint.toPoint()
	}

	// Check if Line 2 is vertical
	if line2End.X-line2Start.X == 0 {

		// Calculate the slope and y-intercept of Line 2
		m1 := (line1End.Y - line1Start.Y) / (line1End.X - line1Start.X)
		b1 := line1Start.Y - m1*line1Start.X

		// Calculate the intersection point
		x := line2Start.X
		y := m1*x + b1

		intersectionPoint := precisePoint{X: x, Y: y}
		return intersectionPoint.toPoint()
	}

	// Calculate the slopes of the lines
	m1 := (line1End.Y - line1Start.Y) / (line1End.X - line1Start.X)
	m2 := (line2End.Y - line2Start.Y) / (line2End.X - line2Start.X)

	// Check if the lines are parallel
	if m1 == m2 {
		return u.PointType{}
	}

	// Calculate the y-intercepts of Line 1 and Line 2
	b1 := line1Start.Y - m1*line1Start.X
	b2 := line2Start.Y - m2*line2Start.X

	// Calculate the intersection point
	x := (b2 - b1) / (m1 - m2)
	y := m1*x + b1

	intersectionPoint := precisePoint{X: x, Y: y}
	return intersectionPoint.toPoint()
}