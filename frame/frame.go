package frame

import (
	u "MM/utils"
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

func NewFrame(poiChan <-chan u.PoiType, pixelDist *u.PixelDistType) *FrameType {
	frame := FrameType{}

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
			log.Infoln("Updated frame, ", poi)
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

	f.guideCorners[cornerNr] = f.Corners[cornerNr]

	switch cornerNr {
	case 0:
		f.guideCorners[0].X += u.GuideCornerOffset
		f.guideCorners[0].Y += u.GuideCornerOffset
	case 1:
		f.guideCorners[1].X -= u.GuideCornerOffset
		f.guideCorners[1].Y += u.GuideCornerOffset
	case 2:
		f.guideCorners[2].X -= u.GuideCornerOffset
		f.guideCorners[2].Y -= u.GuideCornerOffset
	case 3:
		f.guideCorners[3].X += u.GuideCornerOffset
		f.guideCorners[3].Y -= u.GuideCornerOffset
	}
}

func (f *FrameType) NextMove(currentPos u.PointType, nextPos u.PointType) (angle int, dist int) {

	angle, dist = currentPos.Dist(nextPos)

	return
}
