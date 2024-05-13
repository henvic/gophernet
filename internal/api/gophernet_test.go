package api

import (
	"math"
	"testing"
)

func TestBurrowVolume(t *testing.T) {
	b := Burrow{
		Width: 2.0,
		Depth: 3.0,
	}

	expectedVolume := 9.42477796076938 // calculated manually
	volume := b.Volume()

	if math.Abs(volume-expectedVolume) > 0.000001 {
		t.Errorf("Expected volume ~%.6f, but got %.6f", expectedVolume, volume)
	}
}
