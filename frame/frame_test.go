package frame

import (
	u "MM/utils"
	"fmt"
	"math"
	"testing"
	"time"
)

const print = true

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

func setupAngeledFrame() *FrameType {

	poiChan := make(chan u.PoiType)
	frame := NewFrame(poiChan)

	width := 980
	height := 720

	middleXCenter := u.PointType{X: width / 2, Y: height / 2, Angle: 20}
	middleXSize := 80

	poiChan <- u.PoiType{Point: u.PointType{X: 0, Y: 0, Angle: 0}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 980, Y: 0, Angle: 1}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 980, Y: 720, Angle: 2}, Category: u.Corner}
	poiChan <- u.PoiType{Point: u.PointType{X: 0, Y: 720, Angle: 3}, Category: u.Corner}

	angleRad := math.Pi / 180 * float64(middleXCenter.Angle)

	// Place middleX corners in the middle of the field rotated by the given angle
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X + int(math.Cos(angleRad)*float64(middleXSize)), Y: middleXCenter.Y + int(math.Sin(angleRad)*float64(middleXSize)), Angle: 0}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X - int(math.Cos(angleRad)*float64(middleXSize)), Y: middleXCenter.Y - int(math.Sin(angleRad)*float64(middleXSize)), Angle: 1}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X - int(math.Sin(angleRad)*float64(middleXSize)), Y: middleXCenter.Y + int(math.Cos(angleRad)*float64(middleXSize)), Angle: 2}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X + int(math.Sin(angleRad)*float64(middleXSize)), Y: middleXCenter.Y - int(math.Cos(angleRad)*float64(middleXSize)), Angle: 3}, Category: u.MiddleXcorner}

	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X + int(math.Cos(angleRad)*float64(middleXSize)), Y: middleXCenter.Y + int(math.Sin(angleRad)*float64(middleXSize)), Angle: 0}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X - int(math.Cos(angleRad)*float64(middleXSize)), Y: middleXCenter.Y - int(math.Sin(angleRad)*float64(middleXSize)), Angle: 1}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X - int(math.Sin(angleRad)*float64(middleXSize)), Y: middleXCenter.Y + int(math.Cos(angleRad)*float64(middleXSize)), Angle: 2}, Category: u.MiddleXcorner}
	poiChan <- u.PoiType{Point: u.PointType{X: middleXCenter.X + int(math.Sin(angleRad)*float64(middleXSize)), Y: middleXCenter.Y - int(math.Cos(angleRad)*float64(middleXSize)), Angle: 3}, Category: u.MiddleXcorner}

	u.CurrentPos.Set(u.PointType{X: 380, Y: 600})

	return frame
}

func TestFrameIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 380, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 450, Y: 700})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{380 200 0} ball} {{285 360 1} waypoint} {{450 700 0} start}]" {
		t.FailNow()
	}

}

func TestFrameDoubleIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 380, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 600, Y: 550})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{380 200 0} ball} {{285 360 2} waypoint} {{491 564 1} waypoint} {{600 550 0} start}]" {
		t.FailNow()
	}

}

func TestFrameNoIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 100, Y: 350}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 520, Y: 700})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{100 350 0} ball} {{520 700 0} start}]" {
		t.FailNow()
	}

}

func TestFrameSpecialCase(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 268, Y: 532}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 760, Y: 242})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{268 532 0} ball} {{491 564 2} waypoint} {{694 361 1} waypoint} {{760 242 0} start}]" {
		t.FailNow()
	}

}

func TestFrameSpecialCase2(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 650, Y: 550}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{650 550 0} ball} {{700 600 0} start}]" {
		t.FailNow()
	}
}

func TestFrameMiddleXBall(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 510, Y: 390}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{510 390 40} ball} {{590 470 225} precise waypoint} {{607 487 225} waypoint} {{700 600 0} start}]" {
		t.FailNow()
	}

}

func TestFrameMiddleXBall2(t *testing.T) {
	frame := setupFrame()

	//Place ball close to a guideCorner
	nextPos := u.PoiType{Point: u.PointType{X: 540, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 200, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)
	fmt.Println(moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
}

func TestFrameMiddleXBallRemoveWaypoint(t *testing.T) {
	frame := setupFrame()

	//Place ball close to a guideCorner
	nextPos := u.PoiType{Point: u.PointType{X: 300, Y: 320}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	fmt.Println(moves)
	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{300 320 0} ball} {{491 564 1} waypoint} {{700 600 0} start}]" {
		t.FailNow()
	}

}

func TestAngledMiddleXDoubleIntersection(t *testing.T) {
	frame := setupAngeledFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 380, Y: 200}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 600, Y: 550})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, t.Name())
	}
	if fmt.Sprint(moves) != "[{{380 200 0} ball} {{298 290 2} waypoint} {{420 551 1} waypoint} {{600 550 0} start}]" {
		t.FailNow()
	}
}

type testType struct {
	name       string
	currentPos u.PointType
	next       u.PoiType
	expected   string
}

var tests = []testType{
	{name: "TMiddleX1",
		currentPos: u.PointType{X: 700, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 510, Y: 390}, Category: u.Ball},
		expected:   "[{{510 390 40} ball} {{590 470 225} precise waypoint} {{607 487 225} waypoint} {{700 600 0} start}]",
	}, {name: "TMiddleX2",
		currentPos: u.PointType{X: 700, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 470, Y: 340}, Category: u.Ball},
		expected:   "[{{470 340 40} ball} {{389 259 45} precise waypoint} {{372 242 45} waypoint} {{285 360 2} waypoint} {{491 564 1} waypoint} {{700 600 0} start}]",
	}, {name: "TMiddleX3",
		currentPos: u.PointType{X: 100, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 470, Y: 340}, Category: u.Ball},
		expected:   "[{{470 340 40} ball} {{389 259 45} precise waypoint} {{372 242 45} waypoint} {{100 600 0} start}]",
	}, {name: "TBorder",
		currentPos: u.PointType{X: 600, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 200, Y: 10}, Category: u.Ball},
		expected:   "[{{200 10 20} ball} {{200 165 20} precise waypoint} {{285 360 2} waypoint} {{491 564 1} waypoint} {{600 600 0} start}]",
	}, {name: "TCorner",
		currentPos: u.PointType{X: 600, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 5, Y: 5}, Category: u.Ball},
		expected:   "[{{5 5 30} ball} {{128 128 30} precise waypoint} {{148 148 30} precise waypoint} {{168 168 30} waypoint} {{285 360 2} waypoint} {{491 564 1} waypoint} {{600 600 0} start}]",
	}, {name: "TMiddleXDump",
		currentPos: u.PointType{X: 100, Y: 600},
		next:       u.PoiType{Point: u.PointType{X: 5, Y: 360}, Category: u.Goal},
		expected:   "[{{5 360 21} goal} {{57 360 0} waypoint} {{110 360 0} waypoint} {{100 600 0} start}]",
	},
}
var currentTest testType

func TestMultiple(t *testing.T) {
	for _, test := range tests {
		currentTest = test
		t.Run(test.name, testSpecific)
	}
}

func testSpecific(t *testing.T) {
	test := currentTest
	frame := setupFrame()

	nextPos := test.next
	u.CurrentPos.Set(test.currentPos)
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(u.CurrentPos.Get(), nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	if print {
		frame.createTestImg(moves, fmt.Sprint("sub/", test.name))
	}
	if fmt.Sprint(moves) != test.expected {
		t.FailNow()
	}
}
