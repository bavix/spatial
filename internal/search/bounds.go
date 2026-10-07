package search

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/sphere"
)

func Contains(bounds geo.Bounds, point geo.Point) bool {
	return sphere.Contains(bounds.South, bounds.West, bounds.North, bounds.East, point.Lat, point.Lon)
}

func Intersects(a, b geo.Bounds) bool {
	if a.South > b.North || b.South > a.North {
		return false
	}

	if a.West > a.East {
		west := geo.Bounds{
			South: a.South,
			West:  a.West,
			North: a.North,
			East:  180,
		}
		east := geo.Bounds{
			South: a.South,
			West:  -180,
			North: a.North,
			East:  a.East,
		}

		return Intersects(west, b) || Intersects(east, b)
	}

	if b.West > b.East {
		return Intersects(b, a)
	}

	if a.West > b.West {
		a, b = b, a
	}

	return a.East >= b.West || a.West == -180 && b.East == 180
}
