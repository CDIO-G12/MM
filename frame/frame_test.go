package frame

import (
	u "MM/utils"
	"reflect"
	"testing"
	"time"
)

func setupFrame() *FrameType {
	poiChan := make(chan u.PoiType)
	frame := NewFrame(poiChan)

	middleXSize := 80

	poiChan <- u.PoiType{Point: u.PointType{X: 0, Y: 0, Angle: 0}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 980, Y: 0, Angle: 1}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 980, Y: 720, Angle: 2}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 0, Y: 720, Angle: 3}, Category: u.Corner}

	poiChan <- u.PoiType{Point: u.PointType{X: (980/2 - middleXSize), Y: (720 / 2), Angle: 0}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980/2 + middleXSize), Y: (720 / 2), Angle: 1}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980 / 2), Y: (720/2 + middleXSize), Angle: 2}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980 / 2), Y: (720/2 - middleXSize), Angle: 3}, Category: u.MiddleXcorner}

	u.CurrentPos.Set(u.PointType{X: 800, Y: 360, Angle: 0})

	time.Sleep(time.Millisecond)

	return frame
}

func TestFrame1(t *testing.T) {
	frame := setupFrame()

	currentPos := u.PointType{X: 100, Y: 250}
	nextPos := u.PoiType{Point: u.PointType{X: 40, Y: 100}, Category: u.Ball}
	moves := []u.PoiType{{Point: currentPos, Category: u.WayPoint}}
	moves = append(moves, frame.CreateMoves(nextPos)...)
	moves = append(moves, nextPos)

	t.Logf("m: %+v", moves)

	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
	frame.createTestImg(moves, "t1", frame.MiddleX[:], currentPos)
}

func TestFrame2(t *testing.T) {
	frame := setupFrame()

	currentPos := u.PointType{X: 100, Y: 250}
	nextPos := u.PoiType{Point: u.PointType{X: 20, Y: 110}, Category: u.Ball}
	moves := []u.PoiType{{Point: currentPos, Category: u.WayPoint}}
	moves = append(moves, frame.CreateMoves(nextPos)...)
	moves = append(moves, nextPos)

	t.Logf("m: %+v", moves)

	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
	frame.createTestImg(moves, "t2", frame.MiddleX[:], currentPos)
}

func TestFrame3(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 200, Y: 100}, Category: u.Ball}
	currentPos := u.PointType{X: 490, Y: 700}
	currentPos.Angle, _ = currentPos.Dist(nextPos.Point)

	moves := []u.PoiType{{Point: currentPos, Category: u.Start}}
	moves = append(moves, frame.CreateMoves(nextPos)...)

	t.Logf("m: %+v", moves)

	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
	frame.createTestImg(moves, "t2", frame.MiddleX[:], currentPos)

}

/*
func createTestImg(points []u.PointType, name string) {
	width := 250
	height := 150

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	colors := []color.RGBA{{100, 200, 200, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}, {100, 200, 200, 0xff}}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	for i, p := range points {
		img.Set(p.X, p.Y, colors[i])
	}

	// Encode as PNG.
	f, err := os.Create(fmt.Sprint(name, ".png"))
	if err != nil {
		log.Fatal(err)
	}
	err = png.Encode(f, img)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
}
*/
