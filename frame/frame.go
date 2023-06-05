package frame

import (
	u "MM/utils"
	"fmt"
	"image"
	"image/color"
	"image/png"
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

func (f *FrameType) CreateMoves(currentPos u.PointType, nextPos u.PoiType) (directions []u.PoiType) {
	//fmt.Println("Create moves:", currentPos, nextPos)
	first := f.findClosestGuidePosition(currentPos)
	last := f.findClosestGuidePosition(nextPos.Point)
	directions = append(directions, u.PoiType{Point: first, Category: u.WayPoint})

	//angle, dist := f.MiddleXPoint().Dist(currentPos)

	/*TODO: Make middle positions

	Check if middle x is in the way, and add more points
	*/

	directions = append(directions, u.PoiType{Point: last, Category: u.WayPoint})
	directions = append(directions, nextPos)

	f.createTestImg([]u.PoiType{{Point: currentPos}, {Point: first}, {Point: last}, nextPos}, "Directions")

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
	f.MU.Lock()
	defer f.MU.Unlock()
	return f.guideCorners
}
