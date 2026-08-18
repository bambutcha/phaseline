package sim

import "testing"

func TestHexToPixelOrigin(t *testing.T) {
	x, y := HexToPixel(0, 0, 1)
	if x != 0 || y != 0 {
		t.Fatalf("origin: got %v,%v", x, y)
	}
}

func TestCubeDistanceNeighbors(t *testing.T) {
	if CubeDistance(Axial{0, 0}, Axial{1, -1}) != 1 {
		t.Fatal("expected distance 1")
	}
}

func TestPixelRoundTrip(t *testing.T) {
	const size = 16.0
	cases := []Axial{{0, 0}, {1, 0}, {0, 1}, {2, -1}, {-3, 4}}
	for _, h := range cases {
		x, y := HexToPixel(h.Q, h.R, size)
		got := PixelToHex(x, y, size)
		if got != h {
			t.Fatalf("%v roundtrip got %v", h, got)
		}
	}
}

func TestNeighborsCount(t *testing.T) {
	if len(Neighbors) != 6 {
		t.Fatalf("want 6 neighbors, got %d", len(Neighbors))
	}
}
