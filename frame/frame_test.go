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
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 50, Angle: 0}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 50, Angle: 1}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 250, Y: 150, Angle: 2}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 50, Y: 150, Angle: 3}, Category: u.Corner}

	time.Sleep(time.Millisecond)

	return frame
}

func TestFrame1(t *testing.T) {
	frame := setupFrame()

	currentPos := u.PointType{X: 100, Y: 250}
	nextPos := u.PointType{X: 40, Y: 100}
	moves := []u.PointType{currentPos}
	moves = append(moves, frame.CreateMoves(currentPos, nextPos)...)
	moves = append(moves, nextPos)

	t.Logf("m: %+v", moves)

	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
	frame.createTestImg(moves, "t1")
}

func TestFrame2(t *testing.T) {
	frame := setupFrame()

	currentPos := u.PointType{X: 100, Y: 250}
	nextPos := u.PointType{X: 20, Y: 110}
	moves := []u.PointType{currentPos}
	moves = append(moves, frame.CreateMoves(currentPos, nextPos)...)
	moves = append(moves, nextPos)

	t.Logf("m: %+v", moves)

	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
	frame.createTestImg(moves, "t2")
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
