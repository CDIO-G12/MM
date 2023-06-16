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
				frame.updateGuideCorners(poi.Point.Angle)

			case u.Corner:
				if poi.Point.Angle >= len(frame.Corners) {
					continue
				}
				frame.MU.Lock()
				frame.Corners[poi.Point.Angle] = poi.Point
				frame.MU.Unlock()

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

	offset := int(u.GuideCornerOffset * u.GetPixelDist())

	f.guideCorners[cornerNr] = f.MiddleX[cornerNr]

	switch cornerNr {
	case 0: // left
		f.guideCorners[0].X -= offset
	case 1: // right
		f.guideCorners[1].X += offset
	case 2: // top
		f.guideCorners[2].Y -= offset
	case 3: // bottom
		f.guideCorners[3].Y += offset
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

func (f *FrameType) createTestImg(points []u.PoiType, name string, middle []u.PointType) {
	currentPos := u.CurrentPos.Get()
	width := 980
	height := 720

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for _, p := range f.guideCorners {
		f.drawCircle(img, p, 2, color.RGBA{100, 100, 0, 0xff})
	}

	for _, p := range middle {
		f.drawCircle(img, p, 5, color.RGBA{255, 0, 0, 0xff})
	}

	f.drawCircle(img, currentPos, 10, color.RGBA{255, 255, 255, 0xff})

	for _, p := range points {
		if p.Category == u.Ball {
			f.drawCircle(img, p.Point, 5, color.RGBA{255, 0, 255, 0xff})
		}
		if p.Category == u.WayPoint {
			f.drawCircle(img, p.Point, 3, color.RGBA{26, 117, 123, 0xff})
		}
		if p.Category == u.PreciseWayPoint {
			f.drawCircle(img, p.Point, 3, color.RGBA{120, 117, 26, 0xff})
		}
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

func (f *FrameType) drawCircle(img *image.RGBA, center u.PointType, radius int, color color.RGBA) {

	for i := center.X - radius; i <= center.X+radius; i++ {
		for j := center.Y - radius; j <= center.Y+radius; j++ {
			img.Set(i, j, color)
		}
	}

}

func (f *FrameType) GetGuideFrame() [4]u.PointType {
	f.MU.RLock()
	defer f.MU.RUnlock()
	return f.guideCorners
}
