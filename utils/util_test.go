package utils

import "testing"

func TestEtc(t *testing.T) {
	p1 := PointType{X: 100, Y: 100}
	p2 := PointType{X: 50, Y: 50}
	p3 := PointType{X: 50, Y: 100}
	a, d := p1.Dist(p2)
	t.Logf("1-2 d: %d, a: %d", d, a)
	a, d = p1.Dist(p3)
	t.Logf("1-3 d: %d, a: %d", d, a)
	a, d = p2.Dist(p3)
	t.Logf("2-3 d: %d, a: %d", d, a)
}
