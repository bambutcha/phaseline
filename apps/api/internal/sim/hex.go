package sim

import (
	"fmt"
	"math"
)

type Axial struct {
	Q int
	R int
}

var Neighbors = []Axial{
	{1, 0},
	{1, -1},
	{0, -1},
	{-1, 0},
	{-1, 1},
	{0, 1},
}

func HexID(q, r int) string {
	return fmt.Sprintf("%d,%d", q, r)
}

func CubeDistance(a, b Axial) int {
	aq, ar := a.Q, a.R
	bq, br := b.Q, b.R
	return (abs(aq-bq) + abs(aq+ar-bq-br) + abs(ar-br)) / 2
}

func HexToPixel(q, r int, size float64) (x, y float64) {
	x = size * (math.Sqrt(3)*float64(q) + math.Sqrt(3)/2*float64(r))
	y = size * (3.0 / 2.0 * float64(r))
	return x, y
}

func PixelToHex(x, y, size float64) Axial {
	q := (math.Sqrt(3)/3*x - 1.0/3.0*y) / size
	r := (2.0 / 3.0 * y) / size
	return cubeRound(q, -q-r, r)
}

func cubeRound(x, y, z float64) Axial {
	rx, ry, rz := math.Round(x), math.Round(y), math.Round(z)
	dx, dy, dz := math.Abs(rx-x), math.Abs(ry-y), math.Abs(rz-z)
	switch {
	case dx > dy && dx > dz:
		rx = -ry - rz
	case dy > dz:
		ry = -rx - rz
	default:
		rz = -rx - ry
	}
	return Axial{Q: int(rx), R: int(rz)}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
