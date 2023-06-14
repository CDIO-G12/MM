package frame

import (
	u "MM/utils"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"sort"
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
	if cornerNr >= len(f.Corners) {
		return
	}
	f.MU.Lock()
	defer f.MU.Unlock()

	offset := int(u.GuideCornerOffset / u.GetPixelDist())

	f.guideCorners[cornerNr] = f.Corners[cornerNr]

	switch cornerNr {
	case 0:
		f.guideCorners[0].X += offset
		f.guideCorners[0].Y += offset
	case 1:
		f.guideCorners[1].X -= offset
		f.guideCorners[1].Y += offset
	case 2:
		f.guideCorners[2].X -= offset
		f.guideCorners[2].Y -= offset
	case 3:
		f.guideCorners[3].X += offset
		f.guideCorners[3].Y -= offset
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
	f.MU.Lock()
	defer f.MU.Unlock()
	up := u.Avg(f.guideCorners[0].Y, f.guideCorners[1].Y)
	left := u.Avg(f.guideCorners[0].X, f.guideCorners[3].X)
	down := u.Avg(f.guideCorners[2].Y, f.guideCorners[3].Y)
	right := u.Avg(f.guideCorners[1].X, f.guideCorners[2].X)

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
		directions = make([]u.PoiType, 3)
		directions[0] = u.PoiType{Point: u.PointType{X: nextPos.Point.X + int(u.MmToGoal*u.GetPixelDist())*2, Y: nextPos.Point.Y}, Category: u.WayPoint}
		directions[1] = u.PoiType{Point: u.PointType{X: nextPos.Point.X + int(u.MmToGoal*u.GetPixelDist()), Y: nextPos.Point.Y}, Category: u.WayPoint}
		directions[2] = nextPos
		return
	}

	//fmt.Println("Create moves:", currentPos, nextPos)
	first := f.findClosestGuidePosition(currentPos)
	last := f.findClosestGuidePosition(nextPos.Point)
	//directions = append(directions, u.PoiType{Point: first, Category: u.WayPoint})

	directions = append(directions, f.createWaypoint(nextPos)...)
	/*TODO: Make middle positions

	Check if middle x is in the way, and add more points
	*/

	directions = append(directions, u.PoiType{Point: last, Category: u.WayPoint})
	directions = append(directions, nextPos)

	f.createTestImg([]u.PoiType{{Point: currentPos}, {Point: first}, {Point: last}, nextPos}, "Directions", f.MiddleX[:], currentPos)

	//directions = append(directions, nextPos)

	return
}

func (f *FrameType) CalculateWaypoint(nextPos u.PoiType) (WayPoints []u.PoiType) {

	currentPos := u.CurrentPos.Get()

	angleX, distX := currentPos.Dist(f.MiddleXPoint())
	angleB, distB := currentPos.Dist(nextPos.Point)
	yB := nextPos.Point.Y

	safeDist := float64(100)

	//Check if the next ball is on the other side of the middle x
	if distX < distB {

		//Check if the ball is clear of the middle x
		if ((angleX + 10) < angleB) && ((angleX - 10) > angleB) {
			return
		}

		if yB > f.MiddleXPoint().Y {
			x := f.MiddleXPoint().X + int(safeDist*math.Cos(float64(angleB)))
			y := f.MiddleXPoint().Y + int(safeDist*math.Sin(float64(angleB)))

			WayPoints = append(WayPoints, u.PoiType{Point: u.PointType{X: x, Y: y}, Category: u.WayPoint})
		} else {
			x := f.MiddleXPoint().X - int(safeDist*math.Cos(float64(angleB)))
			y := f.MiddleXPoint().Y - int(safeDist*math.Sin(float64(angleB)))

			WayPoints = append(WayPoints, u.PoiType{Point: u.PointType{X: x, Y: y}, Category: u.WayPoint})

		}
	}

	return
}

func (f *FrameType) createWaypoint(nextPos u.PoiType) (WayPoints []u.PoiType) {

	currentPos := u.CurrentPos.Get()

	//safeAngle := 5
	angles := f.findThreeClosestXCorners(currentPos)

	sort.Ints(angles)

	_, distX := currentPos.Dist(f.MiddleXPoint())
	angleB, distB := currentPos.Dist(nextPos.Point)

	max_diff := angles[1] - angles[0]
	p1, p2 := 0, 0

	for i := 0; i < len(angles); i++ {
		for j := 0; j < len(angles); j++ {
			if i == j {
				continue
			}
			diff := u.DegreeSub(angles[j], angles[i])

			if diff > max_diff {
				max_diff = diff
				p1 = i
				p2 = j
			}
		}
	}

	if distB < distX {
		return
	}

	offset := int(-0.00005*math.Pow(float64(distX), 2) + 10)
	//offset := int(15.7 * math.Exp(-0.0004*float64(distX)))

	if u.DegreeSub(angleB, angles[p1])+u.DegreeSub(angleB, angles[p2])+offset > max_diff {

		WayPoints = append(WayPoints, u.PoiType{Point: u.PointType{X: 200, Y: 200, Angle: offset}, Category: u.WayPoint})

		return
	}

	WayPoints = append(WayPoints, u.PoiType{Point: u.PointType{X: 200, Y: 200, Angle: max_diff}, Category: u.WayPoint})

	return

}

func (f *FrameType) findThreeClosestXCorners(pos u.PointType) (returnAngle []int) {

	currentPos := pos

	xPoint := f.MiddleX[:4]

	var angle int

	for _, point := range xPoint {
		angle, _ = currentPos.Dist(point)
		returnAngle = append(returnAngle, angle)
	}

	return
}

const ratingEasy = 0
const ratingHard = 1
const ratingBorder = 2
const ratingCorner = 2
const hardDist = 75
const borderDist = 25
const cornerDist = 25

func (f *FrameType) RateBall(ball *u.PointType) {
	f.MU.RLock()
	defer f.MU.RUnlock()
	pd := u.GetPixelDist()

	for _, corner := range f.MiddleX {
		_, dist := corner.Dist(*ball)
		if dist < int(cornerDist*pd) {
			ball.Angle = ratingCorner
			return
		}
	}

	for _, corner := range f.Corners {
		_, dist := corner.Dist(*ball)
		if dist < cornerDist {
			ball.Angle = ratingCorner
			return
		}
	}

	up := u.Avg(f.Corners[0].Y, f.Corners[1].Y)
	left := u.Avg(f.Corners[0].X, f.Corners[3].X)
	down := u.Avg(f.Corners[2].Y, f.Corners[3].Y)
	right := u.Avg(f.Corners[1].X, f.Corners[2].X)

	bordersDist := []int{u.Abs(ball.Y - up), u.Abs(ball.X - left), u.Abs(ball.Y - down), u.Abs(ball.X - right)}
	min := 999999
	for _, border := range bordersDist {
		if border < min {
			min = border
		}
	}

	if min < borderDist {
		ball.Angle = ratingBorder
	} else if min < hardDist {
		ball.Angle = ratingHard
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

func (f *FrameType) createTestImg(points []u.PoiType, name string, middle []u.PointType, currentPos u.PointType) {
	width := 980
	height := 720

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	//colors := []color.RGBA{{255, 0, 0, 0xff}, {0, 255, 0, 0xff}, {0, 0, 255, 0xff}, {255, 0, 255, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	/*
		for i := f.guideCorners[0].X; i <= f.guideCorners[1].X; i++ {
			img.Set(i, f.guideCorners[0].Y, color.RGBA{200, 200, 200, 0x7F})
			img.Set(i, f.guideCorners[2].Y, color.RGBA{200, 200, 200, 0x7F})
		}

		for i := f.guideCorners[0].Y; i <= f.guideCorners[3].Y; i++ {
			img.Set(f.guideCorners[0].X, i, color.RGBA{200, 200, 200, 0x7F})
			img.Set(f.guideCorners[1].X, i, color.RGBA{200, 200, 200, 0x7F})
		}
	*/

	for _, p := range middle {
		f.drawCircle(img, p, 5)
	}

	img.Set(currentPos.X, currentPos.Y, color.RGBA{255, 255, 255, 0xff})

	for _, p := range points {
		if p.Category == u.Ball {
			f.drawCircle(img, p.Point, 5)
		}
	}

	f.drawCircle(img, currentPos, 5)

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

func (f *FrameType) drawCircle(img *image.RGBA, center u.PointType, radius int) {
	min := 0
	max := 255

	rand1 := uint8(rand.Intn(max-min+1) + min)
	rand2 := uint8(rand.Intn(max-min+1) + min)
	rand3 := uint8(rand.Intn(max-min+1) + min)

	for i := center.X - radius; i <= center.X+radius; i++ {
		for j := center.Y - radius; j <= center.Y+radius; j++ {
			img.Set(i, j, color.RGBA{rand1, rand2, rand3, 0xff})
		}
	}

}

func (f *FrameType) GetGuideFrame() [4]u.PointType {
	f.MU.Lock()
	defer f.MU.Unlock()
	return f.guideCorners
}
