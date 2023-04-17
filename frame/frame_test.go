package frame

import (
	u "MM/utils"
	"testing"
)

func TestFrame(t *testing.T) {
	poiChan := make(chan u.PoiType)
	frame := NewFrame(poiChan, &u.PixelDistType{Definition: 0.10})
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 50, Angle: 0}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 50, Angle: 1}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 250, Angle: 2}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 250, Angle: 3}, Category: u.Corner}

	t.Logf("%+v", frame)

	t.Fail()
}
