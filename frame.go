package main

import (
	"sync"

	log "github.com/s00500/env_logger"
)

type frameType struct {
	corners [4]pointType
	middleX [4]pointType
	goal    pointType
	MU      sync.RWMutex
}

func newFrame(poiChan <-chan poiType) *frameType {
	frame := frameType{}

	go func() {
		for poi := range poiChan {
			switch poi.category {
			case goal:
				frame.MU.Lock()
				frame.goal = poi.point
				frame.MU.Unlock()

			case middleXcorner:
				frame.MU.Lock()
				frame.middleX[poi.point.angle] = poi.point
				frame.MU.Unlock()

			case corner:
				frame.MU.Lock()
				frame.corners[poi.point.angle] = poi.point
				frame.MU.Unlock()

			default:
				continue
			}
			log.Infoln("Updated frame, ", poi)
		}
	}()

	return &frame
}

func (f *frameType) nextMove(currentPos pointType, nextPos pointType) (angle int, dist int) {

	angle, dist = currentPos.dist(nextPos)

	return
}
