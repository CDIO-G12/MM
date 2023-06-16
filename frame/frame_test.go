package frame

import (
	u "MM/utils"
	"fmt"
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

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	frame.createTestImg(moves, "tI")
	if fmt.Sprint(moves) != "[{{380 200 0} ball} {{285 360 1} waypoint} {{450 700 0} start}]" {
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

	t.Log("m:", moves)

	frame.createTestImg(moves, "tDI")
	if fmt.Sprint(moves) != "[{{380 200 0} ball} {{285 360 2} waypoint} {{491 564 1} waypoint} {{600 550 0} start}]" {
		t.FailNow()
	}

}

func TestFrameNoIntersect(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 100, Y: 350}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 520, Y: 700})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	frame.createTestImg(moves, "tNI")
	if fmt.Sprint(moves) != "[{{100 350 0} ball} {{520 700 0} start}]" {
		t.FailNow()
	}

}

func TestFrameSpecialCase(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 268, Y: 532}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 760, Y: 242})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	frame.createTestImg(moves, "tSC")
	if fmt.Sprint(moves) != "[{{268 532 0} ball} {{491 564 2} waypoint} {{694 361 1} waypoint} {{760 242 0} start}]" {
		t.FailNow()
	}

}

func TestFrameSpecialCase2(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 650, Y: 550}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	frame.createTestImg(moves, "tSC2")
	if fmt.Sprint(moves) != "[{{650 550 0} ball} {{700 600 0} start}]" {
		t.FailNow()
	}
}

func TestFrameMiddleXBall(t *testing.T) {
	frame := setupFrame()

	nextPos := u.PoiType{Point: u.PointType{X: 510, Y: 390}, Category: u.Ball}
	u.CurrentPos.Set(u.PointType{X: 700, Y: 600})
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	t.Log("m:", moves)

	frame.createTestImg(moves, "tMX")
	if fmt.Sprint(moves) != "[{{510 390 40} ball} {{767 647 225} precise waypoint} {{870 750 225} waypoint} {{700 600 0} start}]" {
		t.FailNow()
	}

}

type testType struct {
	name       string
	currentPos u.PointType
	ball       u.PointType
	expected   string
}

var tests = []testType{
	{name: "TMiddleX1",
		currentPos: u.PointType{X: 700, Y: 600},
		ball:       u.PointType{X: 510, Y: 390},
		expected:   "[{{510 390 40} ball} {{767 647 225} precise waypoint} {{870 750 225} waypoint} {{700 600 0} start}]",
	}, {name: "TMiddleX2",
		currentPos: u.PointType{X: 700, Y: 600},
		ball:       u.PointType{X: 470, Y: 340},
		expected:   "[{{470 340 40} ball} {{212 82 45} precise waypoint} {{109 -21 45} waypoint} {{285 360 2} waypoint} {{491 564 1} waypoint} {{700 600 0} start}]",
	}, {name: "TMiddleX2",
		currentPos: u.PointType{X: 700, Y: 600},
		ball:       u.PointType{X: 470, Y: 340},
		expected:   "[{{470 340 40} ball} {{212 82 45} precise waypoint} {{109 -21 45} waypoint} {{285 360 2} waypoint} {{491 564 1} waypoint} {{700 600 0} start}]",
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

	nextPos := u.PoiType{Point: test.ball, Category: u.Ball}
	u.CurrentPos.Set(test.currentPos)
	frame.RateBall(&nextPos.Point)

	moves := frame.CreateMoves(nextPos)
	moves = append(moves, u.PoiType{Point: u.CurrentPos.Get(), Category: u.Start})

	fmt.Println("m:", moves)

	go frame.createTestImg(moves, fmt.Sprint("sub/", test.name))
	if fmt.Sprint(moves) != test.expected {
		t.FailNow()
	}
}
