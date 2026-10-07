package metric

import (
	"github.com/bavix/spatial/geo"
	"github.com/bavix/spatial/internal/sphere"
)

type Sphere struct{}

func (Sphere) Distance(a, b geo.Point) float64 {
	return geo.Distance(a, b)
}

func (Sphere) BoundsAround(p geo.Point, r float64) geo.Bounds {
	return geo.BoundsAround(p, r)
}

func (Sphere) LowerBound(p geo.Point, b geo.Bounds) float64 {
	return LowerBound(p, b)
}

func LowerBound(p geo.Point, b geo.Bounds) float64 {
	if !p.Valid() || !b.Valid() {
		return 0
	}

	return sphere.Prepare(p.Lat, p.Lon).LowerBound(b.South, b.West, b.North, b.East)
}
