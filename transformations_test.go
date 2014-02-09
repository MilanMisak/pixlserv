package main

import (
	"image"
	"testing"
)

func TestCalculateTopLeftPointFromGravity(t *testing.T) {
	exp := image.Point{200, 0}
	act := calculateTopLeftPointFromGravity(GRAVITY_NORTH, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("N failed", act, exp)
	}

	exp = image.Point{400, 0}
	act = calculateTopLeftPointFromGravity(GRAVITY_NORTH_EAST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("NE failed", act, exp)
	}

	exp = image.Point{400, 150}
	act = calculateTopLeftPointFromGravity(GRAVITY_EAST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("E failed", act, exp)
	}

	exp = image.Point{400, 300}
	act = calculateTopLeftPointFromGravity(GRAVITY_SOUTH_EAST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("SE failed", act, exp)
	}
	exp = image.Point{200, 300}
	act = calculateTopLeftPointFromGravity(GRAVITY_SOUTH, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("S failed", act, exp)
	}
	exp = image.Point{0, 300}
	act = calculateTopLeftPointFromGravity(GRAVITY_SOUTH_WEST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("SW failed", act, exp)
	}
	exp = image.Point{0, 150}
	act = calculateTopLeftPointFromGravity(GRAVITY_WEST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("W failed", act, exp)
	}

	exp = image.Point{0, 0}
	act = calculateTopLeftPointFromGravity(GRAVITY_NORTH_WEST, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("NW failed", act, exp)
	}

	exp = image.Point{200, 150}
	act = calculateTopLeftPointFromGravity(GRAVITY_CENTER, 400, 300, 800, 600)
	if act != exp {
		t.Errorf("C failed", act, exp)
	}
}
