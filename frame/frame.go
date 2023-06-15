package frame

import (
	u "MM/utils"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"

	log "github.com/s00500/env_logger"
)

type FrameType struct {
	Corners [4]u.PointType
	MiddleX [4]u.PointType
	Goal    u.PointType
	MU      sync.RWMutex

	guideCorners [4]u.PointType
}

func NewFrame(poiChan <-chan u.PoiType) *FrameType {
	frame := FrameType{Corners: [4]u.PointType{{X: 25, Y: 25}}}
	log.Infoln("Frame created")

	go func() {
		for poi := range poiChan {
			switch poi.Category {
			case u.Goal:
				frame.MU.Lock()
				frame.Goal = poi.Point
				frame.MU.Unlock()

			case u.MiddleXcorner:
				frame.MU.Lock()
				frame.MiddleX[poi.Point.Angle] = poi.Point
				frame.MU.Unlock()

			case u.Corner:
				if poi.Point.Angle >= len(frame.Corners) {
					continue
				}
				frame.MU.Lock()
				frame.Corners[poi.Point.Angle] = poi.Point
				frame.MU.Unlock()
				frame.updateGuideCorners(poi.Point.Angle)

			default:
				continue
			}
			//log.Infoln("Updated frame, ", poi)
		}
	}()

	return &frame
}

func (f *FrameType) updateGuideCorners(cornerNr int) {
	if cornerNr >= len(f.MiddleX) {
		return
	}
	f.MU.Lock()
	defer f.MU.Unlock()

	offset := int(u.GuideCornerOffset / u.GetPixelDist())

	f.guideCorners[cornerNr] = f.MiddleX[cornerNr]

	switch cornerNr {
	case 0: // left
		f.MiddleX[0].X -= offset
	case 1: // right
		f.MiddleX[1].X += offset
	case 2: // top
		f.MiddleX[2].Y -= offset
	case 3: // bottom
		f.MiddleX[3].Y += offset
	}
}

func (f *FrameType) MiddleXPoint() u.PointType {
	f.MU.RLock()
	defer f.MU.RUnlock()

	sumX := f.MiddleX[0].X + f.MiddleX[1].X + f.MiddleX[2].X + f.MiddleX[3].X
	sumY := f.MiddleX[0].Y + f.MiddleX[1].Y + f.MiddleX[2].Y + f.MiddleX[3].Y

	middleX := sumX / 4
	middleY := sumY / 4

	return u.PointType{X: middleX, Y: middleY}
}

func (f *FrameType) findClosestGuidePosition(position u.PointType) u.PointType {
	f.MU.RLock()
	up := u.Avg(f.guideCorners[0].Y, f.guideCorners[1].Y)
	left := u.Avg(f.guideCorners[0].X, f.guideCorners[3].X)
	down := u.Avg(f.guideCorners[2].Y, f.guideCorners[3].Y)
	right := u.Avg(f.guideCorners[1].X, f.guideCorners[2].X)

	f.MU.RUnlock()
	bordersDist := []int{u.Abs(position.Y - up), u.Abs(position.X - left), u.Abs(position.Y - down), u.Abs(position.X - right)}
	//borders := []int{up, left, down, right}
	min := 99999
	minI := 5
	for i, border := range bordersDist {
		if border < min {
			min = border
			minI = i
		}
	}

	pos := u.PointType{}
	switch minI {
	case 0: //up
		pos = u.PointType{X: position.X, Y: up}
	case 1: //left
		pos = u.PointType{Y: position.Y, X: left}
	case 2: //down
		pos = u.PointType{X: position.X, Y: down}
	case 3: //right
		pos = u.PointType{Y: position.Y, X: right}
	}

	if pos.X < left {
		pos.X = left
	} else if pos.X > right {
		pos.X = right
	}

	//fmt.Println(pos.Y, up, down)
	if pos.Y < up {
		pos.Y = up
	} else if pos.Y > down {
		pos.Y = down
	}

	return pos
}

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

	directions = append(directions, nextPos)

	if nextPos.Category == u.Ball {

		if nextPos.Point.Angle >= u.RatingCorner {
			corner := nextPos.Point.Angle - u.RatingCorner
			point := nextPos.Point
			offset := int(u.DistanceFromBallCorner * u.GetPixelDist())
			fmt.Println("Corner ", corner, point, offset)
			switch corner {
			case 0: //UpLeft
				point.X += offset
				point.Y += offset
			case 1: //UpRight
				point.X -= offset
				point.Y += offset
			case 2: //DownRight
				point.X -= offset
				point.Y -= offset
			case 3: //DownLeft
				point.X += offset
				point.Y -= offset
			}
			directions = append(directions, u.PoiType{Point: point, Category: u.PreciseWayPoint})
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
			directions = append(directions, u.PoiType{Point: point, Category: u.PreciseWayPoint})
		}
	}

	first := f.findClosestGuidePosition(currentPos)
	last := f.findClosestGuidePosition(nextPos.Point)

	directions = append(directions, f.CalculateWaypoint(nextPos)...)
	/*TODO: Make middle positions

	Check if middle x is in the way, and add more points
	*/

	f.createTestImg([]u.PoiType{{Point: currentPos}, {Point: first}, {Point: last}, nextPos}, "Directions")

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
		ball.Angle = u.RatingCorner + 5
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

func ManualTest() {
	/*poiChan := make(chan u.PoiType)
	frame := NewFrame(poiChan)
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 50, Angle: 0}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 50, Angle: 1}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 150, Angle: 2}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 150, Angle: 3}, Category: u.Corner}

	time.Sleep(time.Millisecond)

	currentPos := u.PointType{X: 110, Y: 70}
	nextPos := u.PointType{X: 220, Y: 110}
	moves := []u.PointType{currentPos}
	moves = append(moves, frame.CreateMoves(currentPos, nextPos)...)
	moves = append(moves, nextPos)

	fmt.Println(moves)
	frame.createTestImg(moves, "t1")*/
}

func (f *FrameType) createTestImg(points []u.PoiType, name string) {
	width := 700
	height := 400

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	colors := []color.RGBA{{255, 0, 0, 0xff}, {0, 255, 0, 0xff}, {0, 0, 255, 0xff}, {255, 0, 255, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for i := f.guideCorners[0].X; i <= f.guideCorners[1].X; i++ {
		img.Set(i, f.guideCorners[0].Y, color.RGBA{200, 200, 200, 0x7F})
		img.Set(i, f.guideCorners[2].Y, color.RGBA{200, 200, 200, 0x7F})
	}

	for i := f.guideCorners[0].Y; i <= f.guideCorners[3].Y; i++ {
		img.Set(f.guideCorners[0].X, i, color.RGBA{200, 200, 200, 0x7F})
		img.Set(f.guideCorners[1].X, i, color.RGBA{200, 200, 200, 0x7F})
	}

	for i, p := range points {
		img.Set(p.Point.X, p.Point.Y, colors[i])
		img.Set(p.Point.X+1, p.Point.Y, colors[i])
		img.Set(p.Point.X+1, p.Point.Y+1, colors[i])
		img.Set(p.Point.X, p.Point.Y+1, colors[i])
	}

	// Encode as PNG.
	file, err := os.Create(fmt.Sprint(name, ".png"))
	if err != nil {
		log.Fatal(err)
	}
	err = png.Encode(file, img)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}

func (f *FrameType) GetGuideFrame() [4]u.PointType {
	f.MU.RLock()
	defer f.MU.RUnlock()
	return f.guideCorners
}
