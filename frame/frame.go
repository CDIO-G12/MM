package frame

import (
	u "MM/utils"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"
	"time"

	log "github.com/s00500/env_logger"
)

type FrameType struct {
	Corners      [4]u.PointType
	MiddleX      [4]u.PointType
	middleXAngle int
	Goal         u.PointType
	mu           sync.Mutex

	guideCorners [4]u.PointType
}

func NewFrame(poiChan <-chan u.PoiType) *FrameType {
	frame := FrameType{Corners: [4]u.PointType{{X: 25, Y: 25}}}

	go func() {
		for poi := range poiChan {
			switch poi.Category {
			case u.Goal:
				frame.mu.Lock()
				frame.Goal = poi.Point
				frame.mu.Unlock()

			case u.MiddleXcorner:
				if poi.Point.Angle >= len(frame.MiddleX) {
					continue
				}
				frame.mu.Lock()
				if frame.MiddleX[poi.Point.Angle].X == 0 || frame.MiddleX[poi.Point.Angle].IsClose(poi.Point, 50) {
					frame.MiddleX[poi.Point.Angle] = poi.Point
				}
				frame.mu.Unlock()
				frame.updateGuideCorners(poi.Point.Angle)

			case u.Corner:
				if poi.Point.Angle >= len(frame.Corners) {
					continue
				}
				frame.mu.Lock()
				frame.Corners[poi.Point.Angle] = poi.Point
				frame.mu.Unlock()

			default:
				continue
			}
			//log.Infoln("Updated frame, ", poi)
		}
	}()

	go func() {
		for {
			time.Sleep(1 * time.Second)
			//frame.mu.Lock()
			frame.createTestImg([]u.PoiType{}, "frame")
			//frame.mu.Unlock()
		}
	}()

	return &frame
}

func (f *FrameType) updateGuideCorners(cornerNr int) {
	if cornerNr >= len(f.MiddleX) {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	offset := int(u.GuideCornerOffset * u.GetPixelDist())

	angleLR, _ := f.MiddleX[1].AngleAndDist(f.MiddleX[0])
	angleUD, _ := f.MiddleX[3].AngleAndDist(f.MiddleX[2])
	angleUD -= 90

	f.middleXAngle = u.Avg(angleLR, angleUD)
	f.middleXAngle = angleLR
	//fmt.Println(f.middleXAngle, angleLR, angleUD)
	/*if f.middleXAngle > 50 {
		f.middleXAngle -= 90
	}*/

	f.guideCorners[cornerNr] = f.MiddleX[cornerNr]
	switch cornerNr {
	case 0: // left
		f.guideCorners[0].Angle = f.middleXAngle
	case 1: // right
		f.guideCorners[1].Angle = f.middleXAngle + 180
	case 2: // top
		f.guideCorners[2].Angle = f.middleXAngle + 90
	case 3: // bottom
		f.guideCorners[3].Angle = f.middleXAngle + 270
	}
	f.guideCorners[cornerNr] = f.guideCorners[cornerNr].CalcNextPos(offset)

}

func (f *FrameType) MiddleXPoint() u.PointType {
	f.mu.Lock()
	sumX := f.MiddleX[0].X + f.MiddleX[1].X + f.MiddleX[2].X + f.MiddleX[3].X
	sumY := f.MiddleX[0].Y + f.MiddleX[1].Y + f.MiddleX[2].Y + f.MiddleX[3].Y
	f.mu.Unlock()

	middleX := sumX / 4
	middleY := sumY / 4

	return u.PointType{X: middleX, Y: middleY}
}

func (f *FrameType) createTestImg(points []u.PoiType, name string) {
	currentPos := u.CurrentPos.Get()
	width := 980
	height := 720

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for _, p := range f.guideCorners {
		f.drawCircle(img, p, 2, color.RGBA{100, 100, 0, 0xff})
	}

	for _, p := range f.MiddleX {
		f.drawCircle(img, p, 5, color.RGBA{255, 0, 0, 0xff})
	}
	f.drawCircle(img, f.MiddleXPoint(), 5, color.RGBA{100, 0, 0, 0xff})

	f.drawCircle(img, currentPos, 10, color.RGBA{255, 255, 255, 0xff})

	for _, p := range points {
		if p.Category == u.Ball {
			f.drawCircle(img, p.Point, 5, color.RGBA{255, 0, 255, 0xff})
		} else if p.Category == u.WayPoint {
			f.drawCircle(img, p.Point, 6, color.RGBA{0, 100, 100, 0xff})
		} else if p.Category == u.PreciseWayPoint {
			f.drawCircle(img, p.Point, 3, color.RGBA{0, 180, 180, 0xff})
		} else if p.Category != u.Start {
			f.drawCircle(img, p.Point, 3, color.RGBA{150, 0, 180, 0xff})
		}
	}

	// Encode as PNG.
	file, err := os.Create(fmt.Sprint("output/", name, ".png"))
	if log.ShouldWarn(err) {
		return
	}
	err = png.Encode(file, img)
	if log.ShouldWarn(err) {
		file.Close()
		return
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
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.guideCorners
}

// sort balls purely based on length to closest
func (f *FrameType) SortBallsOld(balls []u.PointType) (sortedBalls []u.PointType) {
	currentPos := u.CurrentPos.Get()
	//Start from the position of the grapper, not the position of the tracking points
	currentPos = currentPos.CalcNextPos(int(u.DistanceFromBall * u.GetPixelDist()))

	origLength := len(balls)
	if origLength < 2 {
		sortedBalls = balls
		return
	}

	for i := 0; i < origLength; i++ {
		minDist := 99999
		minI := 0
		for j, v := range balls {
			len := currentPos.Dist(v)
			if len < minDist {
				minDist = len
				minI = j
			}
		}
		sortedBalls = append(sortedBalls, balls[minI])
		currentPos = balls[minI]
		balls = u.Remove(balls, minI)
	}
	return
}

// sort balls purely based on length to closest
func (f *FrameType) SortBalls(balls []u.PointType) (sortedBalls []u.PointType) {
	currentPos := u.CurrentPos.Get()
	//Start from the position of the grapper, not the position of the tracking points
	//currentPos = currentPos.CalcNextPos(int(u.DistanceFromBall * u.GetPixelDist()))

	origLength := len(balls)
	// if there is only one ball, return that ball
	if origLength < 2 {
		sortedBalls = balls
		return
	}

	/*
		//uncomment to only sort the 5 closest balls
		if origLength > 5 {
			origLength = 5
		}
	*/

	for i := 0; i < origLength; i++ {
		minDist := 999999
		minI := 0
		for j, v := range balls {
			f.RateBall(&v)
			moves := f.CreateMoves(currentPos, u.PoiType{Point: v, Category: u.Ball})
			dist := 0
			pos := currentPos
			for len(moves) > 0 {
				move := u.Pop(&moves)
				dist += pos.Dist(move.Point)
				pos = move.Point
			}
			// cornerballs should be the last balls
			dist += 500 * v.Angle

			if dist < minDist {
				minDist = dist
				minI = j
			}
		}
		sortedBalls = append(sortedBalls, balls[minI])
		currentPos = balls[minI]
		balls = u.Remove(balls, minI)
	}
	return
}
