package geometry

import (
	"cmp"
	"math"
	"math/big"

	"github.com/bavix/spatial/internal/sphere"
)

const (
	outside = iota
	inside
	boundary
)

const orientationMargin = (3 + 16*0x1p-53) * 0x1p-53

type vertex struct {
	x, y float64
}

func orientation(a, b, c vertex) int {
	if a.x == b.x {
		return cmp.Compare(b.y, a.y) * cmp.Compare(a.x, c.x)
	}

	if a.y == b.y {
		return cmp.Compare(b.x, a.x) * cmp.Compare(c.y, a.y)
	}

	left, right := (a.x-c.x)*(b.y-c.y), (a.y-c.y)*(b.x-c.x)

	det := left - right
	if math.Abs(det) > orientationMargin*(math.Abs(left)+math.Abs(right)) {
		if det > 0 {
			return 1
		}

		return -1
	}

	return exactOrientation(a, b, c)
}

func exactOrientation(a, b, c vertex) int {
	var ax, ay, bx, by, cx, cy, left, right big.Rat
	ax.SetFloat64(a.x)
	ay.SetFloat64(a.y)
	bx.SetFloat64(b.x)
	by.SetFloat64(b.y)
	cx.SetFloat64(c.x)
	cy.SetFloat64(c.y)
	ax.Sub(&ax, &cx)
	ay.Sub(&ay, &cy)
	bx.Sub(&bx, &cx)
	by.Sub(&by, &cy)
	left.Mul(&ax, &by)
	right.Mul(&ay, &bx)

	return left.Sub(&left, &right).Sign()
}

func onSegment(a, b, p vertex) bool {
	return p.x >= min(a.x, b.x) && p.x <= max(a.x, b.x) && p.y >= min(a.y, b.y) && p.y <= max(a.y, b.y) && orientation(a, b, p) == 0
}

func intersects(a, b, c, d vertex) bool {
	abc, abd, cda, cdb := orientation(a, b, c), orientation(a, b, d), orientation(c, d, a), orientation(c, d, b)

	return abc*abd < 0 && cda*cdb < 0 ||
		abc == 0 && onSegment(a, b, c) || abd == 0 && onSegment(a, b, d) ||
		cda == 0 && onSegment(c, d, a) || cdb == 0 && onSegment(c, d, b)
}

func cross(a, b, p vertex, location int) int {
	if onSegment(a, b, p) {
		return boundary
	}

	if (a.y > p.y) != (b.y > p.y) && (orientation(a, b, p) > 0) == (b.y > a.y) {
		return inside - location
	}

	return location
}

func locate(r Ring, p Point) int {
	if !p.Valid() || len(r) < 3 || !r[0].Valid() {
		return outside
	}

	anchor := r[0].Lon
	point := vertex{
		anchor + sphere.LongitudeDifference(anchor, p.Lon),
		p.Lat,
	}
	first := vertex{
		anchor,
		r[0].Lat,
	}
	previous := first
	location := outside

	for row := 1; row < len(r); row++ {
		if !r[row].Valid() {
			return outside
		}

		current := vertex{
			previous.x + sphere.LongitudeDifference(r[row-1].Lon, r[row].Lon),
			r[row].Lat,
		}

		location = cross(previous, current, point, location)
		if location == boundary {
			return boundary
		}

		previous = current
	}

	return cross(previous, first, point, location)
}
