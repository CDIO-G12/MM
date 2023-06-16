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
	poiChan <- u.PoiType{Point: u.PointType{X: (980/2 + middleXSize), Y: (722 / 2), Angle: 1}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980 / 2), Y: (720/2 - middleXSize), Angle: 2}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (982 / 2), Y: (720/2 + middleXSize), Angle: 3}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980/2 - middleXSize), Y: (720 / 2), Angle: 0}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980/2 + middleXSize), Y: (722 / 2), Angle: 1}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (980 / 2), Y: (720/2 - middleXSize), Angle: 2}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: (982 / 2), Y: (720/2 + middleXSize), Angle: 3}, Category: u.MiddleXcorner}
	time.Sleep(time.Millisecond)

	u.CurrentPos.Set(u.PointType{X: 380, Y: 600})

	return frame
}

func TestFrameIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 380, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 450, Y: 700})
	frame.RateBall(&nextPos.Point)

	moves := []u.PoiType{{Point: u.CurrentPos.Get(), Category: u.Start}}
	moves = append(moves, frame.CreateMoves(nextPos)...)

	t.Logf("m: %+v", moves)

	frame.createTestImg(moves, "tI", frame.MiddleX[:])
	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}

}

func TestFrameDoubleIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 380, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 600, Y: 550})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Logf("m: %+v", moves)

	frame.createTestImg(moves, "tDI", frame.MiddleX[:])
	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}

}

func TestFrameNoIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 100, Y: 350}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 520, Y: 700})
	frame.RateBall(&nextPos.Point)

	moves := []u.PoiType{{Point: u.CurrentPos.Get(), Category: u.Start}}
	moves = append(moves, frame.CreateMoves(nextPos)...)

	t.Logf("m: %+v", moves)

	frame.createTestImg(moves, "tNI", frame.MiddleX[:])
	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}

}

func TestFrameSpecialCase(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 268, Y: 532}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 760, Y: 242})
	frame.RateBall(&nextPos.Point)

	moves := []u.PoiType{{Point: u.CurrentPos.Get(), Category: u.Start}}
	moves = append(moves, frame.CreateMoves(nextPos)...)

	t.Logf("m: %+v", moves)

	frame.createTestImg(moves, "tSC", frame.MiddleX[:])
	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}

}

func TestFrameSpecialCase2(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 650, Y: 550}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := []u.PoiType{{Point: u.CurrentPos.Get(), Category: u.Start}}
	moves = append(moves, frame.CreateMoves(nextPos)...)

	t.Logf("m: %+v", moves)

	frame.createTestImg(moves, "tSC2", frame.MiddleX[:])
	if !reflect.DeepEqual(moves, []u.PointType{{X: 100, Y: 240}, {X: 60, Y: 100}}) {
		t.FailNow()
	}
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
